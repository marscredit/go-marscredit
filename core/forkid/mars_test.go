package forkid

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

func TestMarsForkID(t *testing.T) {
	z := big.NewInt(0)
	cfg := &params.ChainConfig{ChainID: big.NewInt(110110), HomesteadBlock: z, EIP150Block: z, EIP155Block: z,
		EIP158Block: z, ByzantiumBlock: z, ConstantinopleBlock: z, PetersburgBlock: z, IstanbulBlock: z,
		BerlinBlock: z, LondonBlock: z, Ethash: new(params.EthashConfig)}
	genesis := common.HexToHash("0x01")
	f := params.MarsFastForkBlock.Uint64()
	before := NewID(cfg, genesis, f-1)
	after := NewID(cfg, genesis, f)
	if before.Next != f || after.Next != 0 || before.Hash == after.Hash {
		t.Fatalf("unexpected fork ids: before %+v after %+v", before, after)
	}
	// pre-fork, upgraded and stock nodes share the fork hash (only FORK_NEXT differs)
	// stock geth gathers no Mars fork; simulate by moving the fork to 0 (dropped)
	old := params.MarsFastForkBlock
	params.MarsFastForkBlock = big.NewInt(0)
	stockID := NewID(cfg, genesis, f-1)
	params.MarsFastForkBlock = old
	if stockID.Hash != before.Hash {
		t.Fatalf("pre-fork hash must match stock nodes: %x vs %x", stockID.Hash, before.Hash)
	}
}
