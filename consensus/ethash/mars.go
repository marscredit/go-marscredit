// Copyright 2026 The Mars Credit Authors
// This file is part of the go-ethereum library (Mars Credit fork).

package ethash

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// errMarsMissingAncestor is returned when the LWMA window cannot be assembled.
var errMarsMissingAncestor = errors.New("mars lwma: missing ancestor header")

// CalcDifficultyMars returns the difficulty of the child of parent under the
// Mars fast-block rules (params.IsMarsFast(parent.Number+1) must hold).
//
//   - fork block: one-time drop, D = D(parent) / MarsForkDifficultyDivisor
//   - fewer than MarsLWMAMinWindow post-fork solvetimes: keep D(parent)
//   - otherwise zawy12 LWMA-1 over the last min(N, available) post-fork blocks:
//     next = avg(D) * w*(w+1)*T*99 / (200 * sum(i * clamp(st_i, -6T, 6T)))
//     with sum floored at w*w*T/20
//   - floor params.MarsMinimumDifficulty; no difficulty bomb.
//
// Only headers strictly after the fork block contribute solvetimes, so the
// pre-fork 12 s history never leaks in. Ancestors are read through chain, which
// in batch verification is an overlay over the batch (see marsBatchReader).
func CalcDifficultyMars(chain consensus.ChainHeaderReader, parent *types.Header) (*big.Int, error) {
	next := new(big.Int).Add(parent.Number, big1)
	fork := params.MarsFastForkBlock
	if next.Cmp(fork) == 0 {
		d := new(big.Int).Div(parent.Difficulty, big.NewInt(params.MarsForkDifficultyDivisor))
		return marsFloor(d), nil
	}
	// first block whose solvetime is used
	n := int64(params.MarsLWMAWindow)
	first := new(big.Int).Sub(next, big.NewInt(n))
	if minFirst := new(big.Int).Add(fork, big1); first.Cmp(minFirst) < 0 {
		first = minFirst
	}
	w := new(big.Int).Sub(next, first).Int64() // number of solvetimes
	if w < params.MarsLWMAMinWindow {
		return new(big.Int).Set(parent.Difficulty), nil
	}
	// hs[0] is block first-1 (timestamp only), hs[1..w] are the window.
	hs := make([]*types.Header, w+1)
	h := parent
	for i := w; i >= 0; i-- {
		hs[i] = h
		if i > 0 {
			h = chain.GetHeader(h.ParentHash, h.Number.Uint64()-1)
			if h == nil {
				return nil, errMarsMissingAncestor
			}
		}
	}
	T := int64(params.MarsTargetBlockTime)
	var L int64
	sumD := new(big.Int)
	for i := int64(1); i <= w; i++ {
		st := int64(hs[i].Time) - int64(hs[i-1].Time)
		if st > 6*T {
			st = 6 * T
		} else if st < -6*T {
			st = -6 * T
		}
		L += i * st
		sumD.Add(sumD, hs[i].Difficulty)
	}
	if minL := w * w * T / 20; L < minL {
		L = minL
	}
	avgD := new(big.Int).Div(sumD, big.NewInt(w))
	num := new(big.Int).Mul(avgD, big.NewInt(w*(w+1)*T*99))
	d := num.Div(num, big.NewInt(200*L))
	return marsFloor(d), nil
}

func marsFloor(d *big.Int) *big.Int {
	if d.Cmp(params.MarsMinimumDifficulty) < 0 {
		return new(big.Int).Set(params.MarsMinimumDifficulty)
	}
	return d
}

// marsBatchReader overlays a batch of not-yet-imported headers on top of the
// chain so LWMA can look back through earlier headers of the same batch.
type marsBatchReader struct {
	consensus.ChainHeaderReader
	byHash map[common.Hash]*types.Header
}

func newMarsBatchReader(chain consensus.ChainHeaderReader, headers []*types.Header) *marsBatchReader {
	m := make(map[common.Hash]*types.Header, len(headers))
	for _, h := range headers {
		m[h.Hash()] = h
	}
	return &marsBatchReader{ChainHeaderReader: chain, byHash: m}
}

func (r *marsBatchReader) GetHeader(hash common.Hash, number uint64) *types.Header {
	if h, ok := r.byHash[hash]; ok && h.Number.Uint64() == number {
		return h
	}
	return r.ChainHeaderReader.GetHeader(hash, number)
}

// GetHeaderByHash is overridden too so the overlay is consistent.
func (r *marsBatchReader) GetHeaderByHash(hash common.Hash) *types.Header {
	if h, ok := r.byHash[hash]; ok {
		return h
	}
	return r.ChainHeaderReader.GetHeaderByHash(hash)
}

// allowedFutureSeconds returns the future-block allowance for header number.
func allowedFutureSeconds(config *params.ChainConfig, number *big.Int) int64 {
	if config.IsMarsFast(number) {
		return params.MarsAllowedFutureBlockTimeSeconds
	}
	return allowedFutureBlockTimeSeconds
}
