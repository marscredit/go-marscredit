package params

import (
	"math/big"
	"strings"
	"testing"
)

func TestMarsForkKeyedByChainID(t *testing.T) {
	f := MarsFastForkBlock
	mars := &ChainConfig{ChainID: big.NewInt(110110)}
	other := &ChainConfig{ChainID: big.NewInt(1)}
	if mars.IsMarsFast(new(big.Int).Sub(f, big.NewInt(1))) || !mars.IsMarsFast(f) {
		t.Fatal("Mars activation boundary wrong")
	}
	if other.IsMarsFast(f) || (&ChainConfig{}).IsMarsFast(f) {
		t.Fatal("fork must only apply to chain 110110")
	}
	if !strings.Contains(mars.String(), "Mars Fast: "+f.String()) || !strings.Contains(other.String(), "Mars Fast: n/a") {
		t.Fatal("String() missing Mars Fast")
	}
	if VersionMeta != "mars-fast1" {
		t.Fatal("VersionMeta")
	}
}
