// Copyright 2026 The Mars Credit Authors
// This file is part of the go-ethereum library (Mars Credit fork).

package params

import "math/big"

// Mars Credit "fast blocks" hard fork (mars-fast1).
//
// The fork is keyed on the chain ID in code, so it activates for chain 110110
// no matter what chain config a node has stored in its database. It changes:
//   - difficulty: zawy12 LWMA-1 (target MarsTargetBlockTime, window
//     MarsLWMAWindow), no difficulty bomb, Mars-only floor MarsMinimumDifficulty
//   - allowed future block time: MarsAllowedFutureBlockTimeSeconds (was 15)
// The block reward is unchanged (stock Constantinople 2 MARS).
var (
	// MarsChainID is the Mars Credit mainnet chain ID.
	MarsChainID = big.NewInt(110110)

	// MarsFastForkBlock is the first block mined under the fast-block rules.
	// PLACEHOLDER: about 24h of blocks above the tip on 2026-10-15 (5,137,880).
	// Must be confirmed before release.
	MarsFastForkBlock = big.NewInt(5_145_000)

	// MarsMinimumDifficulty is the post-fork difficulty floor (stock 131072).
	MarsMinimumDifficulty = big.NewInt(16384)
)

const (
	// MarsTargetBlockTime is the post-fork target block time in seconds.
	MarsTargetBlockTime = 2
	// MarsLWMAWindow is the LWMA-1 averaging window N, in blocks.
	MarsLWMAWindow = 120
	// MarsLWMAMinWindow is how many post-fork blocks must exist before LWMA
	// replaces the fork-block bootstrap difficulty.
	MarsLWMAMinWindow = 10
	// MarsForkDifficultyDivisor sets the one-time difficulty drop at the fork
	// block: D(fork) = D(parent) / divisor (12.2 s -> ~2 s at constant hashrate).
	MarsForkDifficultyDivisor = 6
	// MarsAllowedFutureBlockTimeSeconds replaces the stock 15 s future-block
	// allowance for post-fork headers.
	MarsAllowedFutureBlockTimeSeconds = int64(2)
	// MarsMinerMinGapMillis is the miner-side wait after a new head is imported
	// before the local miner starts sealing on top of it (not a consensus rule).
	MarsMinerMinGapMillis = 1000
)

// IsMarsChain reports whether the config is the Mars Credit chain.
func (c *ChainConfig) IsMarsChain() bool {
	return c != nil && c.ChainID != nil && c.ChainID.Cmp(MarsChainID) == 0
}

// IsMarsFast reports whether block num runs under the Mars fast-block rules.
func (c *ChainConfig) IsMarsFast(num *big.Int) bool {
	return c.IsMarsChain() && isForked(MarsFastForkBlock, num)
}
