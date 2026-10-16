// Copyright 2026 The Mars Credit Authors
// This file is part of the go-ethereum library (Mars Credit fork).

package ethash

import (
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// marsConfig mirrors the Mars Credit genesis (miner-apple-silicon
// Resources/mars_credit_genesis.json): every fork through London at block 0.
func marsConfig() *params.ChainConfig {
	z := big.NewInt(0)
	return &params.ChainConfig{
		ChainID: big.NewInt(110110), HomesteadBlock: z, EIP150Block: z, EIP155Block: z, EIP158Block: z,
		ByzantiumBlock: z, ConstantinopleBlock: z, PetersburgBlock: z, IstanbulBlock: z, BerlinBlock: z,
		LondonBlock: z, Ethash: new(params.EthashConfig),
	}
}

type marsTestChain struct {
	config  *params.ChainConfig
	headers map[common.Hash]*types.Header
	byNum   map[uint64]*types.Header
}

func newMarsTestChain(cfg *params.ChainConfig, hs ...*types.Header) *marsTestChain {
	c := &marsTestChain{config: cfg, headers: map[common.Hash]*types.Header{}, byNum: map[uint64]*types.Header{}}
	for _, h := range hs {
		c.add(h)
	}
	return c
}
func (c *marsTestChain) add(h *types.Header)          { c.headers[h.Hash()] = h; c.byNum[h.Number.Uint64()] = h }
func (c *marsTestChain) Config() *params.ChainConfig  { return c.config }
func (c *marsTestChain) CurrentHeader() *types.Header { return nil }
func (c *marsTestChain) GetHeader(hash common.Hash, n uint64) *types.Header {
	if h := c.headers[hash]; h != nil && h.Number.Uint64() == n {
		return h
	}
	return nil
}
func (c *marsTestChain) GetHeaderByNumber(n uint64) *types.Header       { return c.byNum[n] }
func (c *marsTestChain) GetHeaderByHash(hash common.Hash) *types.Header { return c.headers[hash] }
func (c *marsTestChain) GetTd(common.Hash, uint64) *big.Int             { return nil }

type rpcHeader struct {
	types.Header
	Hash common.Hash `json:"hash"`
}

func loadRealMarsHeaders(t *testing.T) []*types.Header {
	raw, err := os.ReadFile("testdata/mars_headers_5137680_5137880.json")
	if err != nil {
		t.Fatal(err)
	}
	var rs []json.RawMessage
	if err := json.Unmarshal(raw, &rs); err != nil {
		t.Fatal(err)
	}
	out := make([]*types.Header, 0, len(rs))
	for _, r := range rs {
		var h types.Header
		if err := json.Unmarshal(r, &h); err != nil {
			t.Fatal(err)
		}
		var hh struct{ Hash common.Hash }
		json.Unmarshal(r, &hh)
		if h.Hash() != hh.Hash {
			t.Fatalf("block %d: rebuilt header hash %x != rpc hash %x", h.Number, h.Hash(), hh.Hash)
		}
		out = append(out, &h)
	}
	return out
}

// Pre-fork difficulty must be exactly stock geth for 200 real Mars blocks.
func TestMarsPreForkRealDifficulty(t *testing.T) {
	cfg := marsConfig()
	hs := loadRealMarsHeaders(t)
	if params.MarsFastForkBlock.Cmp(hs[len(hs)-1].Number) <= 0 {
		t.Fatalf("fixture must be pre-fork")
	}
	chain := newMarsTestChain(cfg, hs...)
	e := NewFaker()
	for i := 1; i < len(hs); i++ {
		want := hs[i].Difficulty
		if got := CalcDifficulty(cfg, hs[i].Time, hs[i-1]); got.Cmp(want) != 0 {
			t.Fatalf("block %d: stock difficulty %v, chain has %v", hs[i].Number, got, want)
		}
		if got := e.CalcDifficulty(chain, hs[i].Time, hs[i-1]); got.Cmp(want) != 0 {
			t.Fatalf("block %d: engine difficulty %v, chain has %v", hs[i].Number, got, want)
		}
	}
}

// Full header verification (incl. EIP-1559 base fee) of real blocks in batch
// mode with only the first header known to the chain. Seals are checked with a
// real ethash light cache unless -short.
func TestMarsRealHeadersVerify(t *testing.T) {
	cfg := marsConfig()
	hs := loadRealMarsHeaders(t)
	chain := newMarsTestChain(cfg, hs[0])
	var e *Ethash
	seals := make([]bool, len(hs)-1)
	if testing.Short() {
		e = NewFaker()
	} else {
		e = New(Config{PowMode: ModeNormal, CacheDir: t.TempDir(), CachesInMem: 1}, nil, false)
		for i := range seals {
			seals[i] = i%20 == 0 // sample of PoW seals (each needs the epoch cache)
		}
	}
	defer e.Close()
	_, results := e.VerifyHeaders(chain, hs[1:], seals)
	for i := range hs[1:] {
		if err := <-results; err != nil {
			t.Fatalf("real block %d failed verification: %v", hs[i+1].Number, err)
		}
	}
	if !testing.Short() {
		// negative control: a corrupted nonce must fail the real PoW check
		bad := types.CopyHeader(hs[1])
		bad.Nonce[7] ^= 0xff
		if err := e.verifySeal(chain, bad, false); err == nil {
			t.Fatalf("corrupted seal accepted")
		}
	}
}

// --- synthetic post-fork chain helpers ---

func withFork(t *testing.T, n int64) {
	old := params.MarsFastForkBlock
	params.MarsFastForkBlock = big.NewInt(n)
	t.Cleanup(func() { params.MarsFastForkBlock = old })
}

// makeMarsChain builds a valid header chain (fake seals) with the given
// solvetimes, computing each difficulty with the engine.
func makeMarsChain(t *testing.T, cfg *params.ChainConfig, chain *marsTestChain, parent *types.Header, solvetimes []uint64) []*types.Header {
	e := NewFaker()
	var out []*types.Header
	for _, st := range solvetimes {
		h := &types.Header{
			ParentHash: parent.Hash(), UncleHash: types.EmptyUncleHash, Number: new(big.Int).Add(parent.Number, big1),
			GasLimit: parent.GasLimit, Time: parent.Time + st, Extra: []byte{},
		}
		h.BaseFee = misc.CalcBaseFee(cfg, parent)
		if err := e.Prepare(chain, h); err != nil {
			t.Fatal(err)
		}
		chain.add(h)
		out = append(out, h)
		parent = h
	}
	return out
}

func preForkBase(fork int64, d int64, t0 uint64) []*types.Header {
	var hs []*types.Header
	var parent common.Hash
	for n := fork - 5; n < fork; n++ {
		h := &types.Header{ParentHash: parent, UncleHash: types.EmptyUncleHash, Number: big.NewInt(n),
			GasLimit: 30_000_000, Time: t0 + uint64(n-(fork-5))*12, Difficulty: big.NewInt(d), BaseFee: big.NewInt(7), Extra: []byte{}}
		parent = h.Hash()
		hs = append(hs, h)
	}
	return hs
}

// reference LWMA-1 in float64, straight from zawy12's formula.
func refLWMA(ts []uint64, ds []*big.Int) float64 {
	w := len(ds)
	T := float64(params.MarsTargetBlockTime)
	L, sum := 0.0, 0.0
	for i := 1; i <= w; i++ {
		st := float64(int64(ts[i]) - int64(ts[i-1]))
		if st > 6*T {
			st = 6 * T
		}
		L += float64(i) * st
		f, _ := new(big.Float).SetInt(ds[i-1]).Float64()
		sum += f
	}
	if L < float64(w*w)*T/20 {
		L = float64(w*w) * T / 20
	}
	return sum / float64(w) * float64(w*(w+1)) * T * 0.99 / (2 * L)
}

func TestMarsLWMA(t *testing.T) {
	const fork = 1000
	withFork(t, fork)
	cfg := marsConfig()
	base := preForkBase(fork, 2_544_000, uint64(time.Now().Unix())-100_000)
	chain := newMarsTestChain(cfg, base...)
	// fork block: one-time drop
	sts := []uint64{12}
	for i := 0; i < 400; i++ {
		sts = append(sts, uint64(1+i%4)) // mean 2.5 s
	}
	hs := makeMarsChain(t, cfg, chain, base[len(base)-1], sts)
	if want := big.NewInt(2_544_000 / 6); hs[0].Difficulty.Cmp(want) != 0 {
		t.Fatalf("fork block difficulty %v want %v", hs[0].Difficulty, want)
	}
	// bootstrap: held until MarsLWMAMinWindow post-fork solvetimes exist
	for i := 1; i <= params.MarsLWMAMinWindow; i++ { // w = i-1 < MinWindow
		if hs[i].Difficulty.Cmp(hs[0].Difficulty) != 0 {
			t.Fatalf("block %d should hold bootstrap difficulty", hs[i].Number)
		}
	}
	// every later block equals the float reference within rounding
	for i := params.MarsLWMAMinWindow + 1; i < len(hs); i++ {
		w := i - 1 // solvetimes from post-fork blocks hs[1..i-1]
		if w > params.MarsLWMAWindow {
			w = params.MarsLWMAWindow
		}
		var ts []uint64
		var ds []*big.Int
		for j := i - w - 1; j < i; j++ {
			ts = append(ts, hs[j].Time)
			if j > i-w-1 {
				ds = append(ds, hs[j].Difficulty)
			}
		}
		ref := refLWMA(ts, ds)
		got, _ := new(big.Float).SetInt(hs[i].Difficulty).Float64()
		if d := got - ref; d > 3 || d < -3 { // integer floor rounding only
			t.Fatalf("block %d: lwma %v ref %.0f", hs[i].Number, hs[i].Difficulty, ref)
		}
	}
	// mean solvetime 2.5 s > target: difficulty must have fallen
	if hs[len(hs)-1].Difficulty.Cmp(hs[0].Difficulty) >= 0 {
		t.Fatalf("difficulty did not fall for slow blocks")
	}
	// fast blocks raise difficulty; floor holds for very slow blocks
	up := makeMarsChain(t, cfg, chain, hs[len(hs)-1], repeat(1, 200))
	if up[len(up)-1].Difficulty.Cmp(hs[len(hs)-1].Difficulty) <= 0 {
		t.Fatalf("difficulty did not rise for 1 s blocks")
	}
	down := makeMarsChain(t, cfg, chain, up[len(up)-1], repeat(500, 400))
	if down[len(down)-1].Difficulty.Cmp(params.MarsMinimumDifficulty) != 0 {
		t.Fatalf("floor not applied: %v", down[len(down)-1].Difficulty)
	}
	// determinism: recompute from scratch on a fresh reader
	chain2 := newMarsTestChain(cfg, base...)
	hs2 := makeMarsChain(t, cfg, chain2, base[len(base)-1], sts)
	for i := range hs {
		if hs[i].Hash() != hs2[i].Hash() {
			t.Fatalf("non-deterministic at %d", i)
		}
	}
	// missing ancestors are an error, not a panic
	orphanParent := hs[300]
	empty := newMarsTestChain(cfg)
	if d := NewFaker().CalcDifficulty(empty, orphanParent.Time+2, orphanParent); d != nil {
		t.Fatalf("expected nil difficulty with missing ancestors")
	}
}

func repeat(v uint64, n int) []uint64 {
	out := make([]uint64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

// A post-fork batch whose LWMA window lies entirely inside the batch must
// verify in VerifyHeaders, while single-header verification against a chain
// that lacks those ancestors cannot.
func TestMarsBatchVerify(t *testing.T) {
	const fork = 1000
	withFork(t, fork)
	cfg := marsConfig()
	base := preForkBase(fork, 2_544_000, uint64(time.Now().Unix())-100_000)
	full := newMarsTestChain(cfg, base...)
	sts := []uint64{12}
	for i := 0; i < 299; i++ {
		sts = append(sts, uint64(1+(i*7)%4))
	}
	hs := makeMarsChain(t, cfg, full, base[len(base)-1], sts)

	known := newMarsTestChain(cfg, base...) // batch headers unknown to the chain
	e := NewFaker()
	_, results := e.VerifyHeaders(known, hs, make([]bool, len(hs)))
	for i := range hs {
		if err := <-results; err != nil {
			t.Fatalf("batch header %d: %v", hs[i].Number, err)
		}
	}
	// single header deep in the batch: ancestors missing from chain
	known.add(hs[249])
	if err := e.VerifyHeader(known, hs[250], false); !errors.Is(err, errMarsMissingAncestor) {
		t.Fatalf("expected missing-ancestor error, got %v", err)
	}
	// tampered difficulty is rejected in batch mode
	bad := make([]*types.Header, len(hs))
	for i, h := range hs {
		bad[i] = types.CopyHeader(h)
	}
	bad[200].Difficulty = new(big.Int).Add(bad[200].Difficulty, big1)
	for i := 201; i < len(bad); i++ {
		bad[i].ParentHash = bad[i-1].Hash()
	}
	_, results = e.VerifyHeaders(newMarsTestChain(cfg, base...), bad, make([]bool, len(bad)))
	for i := range bad {
		err := <-results
		if i == 200 && err == nil {
			t.Fatalf("tampered difficulty accepted")
		}
		if i < 200 && err != nil {
			t.Fatalf("header %d: %v", i, err)
		}
	}
}

func TestMarsFutureBlockTime(t *testing.T) {
	const fork = 1000
	withFork(t, fork)
	cfg := marsConfig()
	now := uint64(time.Now().Unix())
	e := NewFaker()

	check := func(parent *types.Header, chain *marsTestChain, ahead uint64) error {
		h := &types.Header{ParentHash: parent.Hash(), UncleHash: types.EmptyUncleHash, Number: new(big.Int).Add(parent.Number, big1),
			GasLimit: parent.GasLimit, Time: now + ahead, Extra: []byte{}, BaseFee: misc.CalcBaseFee(cfg, parent)}
		if err := e.Prepare(chain, h); err != nil {
			t.Fatal(err)
		}
		return e.VerifyHeader(chain, h, false)
	}
	// pre-fork: stock 15 s
	pre := preForkBase(fork-10, 2_544_000, now-1000)
	preChain := newMarsTestChain(cfg, pre...)
	if err := check(pre[len(pre)-1], preChain, 10); err != nil {
		t.Fatalf("pre-fork +10s should pass: %v", err)
	}
	if err := check(pre[len(pre)-1], preChain, 20); !errors.Is(err, consensus.ErrFutureBlock) {
		t.Fatalf("pre-fork +20s should be future: %v", err)
	}
	// post-fork: 2 s
	base := preForkBase(fork, 2_544_000, now-1000)
	chain := newMarsTestChain(cfg, base...)
	hs := makeMarsChain(t, cfg, chain, base[len(base)-1], repeat(2, 50))
	parent := hs[len(hs)-1]
	if err := check(parent, chain, 0); err != nil {
		t.Fatalf("post-fork now should pass: %v", err)
	}
	if err := check(parent, chain, 4); !errors.Is(err, consensus.ErrFutureBlock) {
		t.Fatalf("post-fork +4s should be future: %v", err)
	}
	// other chains keep 15 s even after the Mars height
	other := marsConfig()
	other.ChainID = big.NewInt(1)
	oc := newMarsTestChain(other, base...)
	if err := check(base[len(base)-1], oc, 10); err != nil {
		t.Fatalf("non-Mars chain +10s should pass: %v", err)
	}
}
