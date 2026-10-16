package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func TestMarsSealDelay(t *testing.T) {
	parent := common.HexToHash("0x01")
	blk := func(n int64) *types.Block {
		return types.NewBlockWithHeader(&types.Header{Number: big.NewInt(n), ParentHash: parent})
	}
	f := params.MarsFastForkBlock.Int64()
	w := &worker{chainConfig: &params.ChainConfig{ChainID: big.NewInt(110110)}}
	w.marsNoteHead(parent)
	if d := w.marsSealDelay(blk(f)); d < 900*time.Millisecond || d > time.Second {
		t.Fatalf("post-fork delay %v, want ~1s", d)
	}
	if d := w.marsSealDelay(blk(f - 1)); d != 0 {
		t.Fatalf("pre-fork delay %v, want 0", d)
	}
	time.Sleep(1100 * time.Millisecond)
	if d := w.marsSealDelay(blk(f)); d > 0 {
		t.Fatalf("delay should have elapsed, got %v", d)
	}
	other := &worker{chainConfig: &params.ChainConfig{ChainID: big.NewInt(1)}}
	other.marsNoteHead(parent)
	if d := other.marsSealDelay(blk(f)); d != 0 {
		t.Fatalf("non-Mars chain delay %v", d)
	}
}
