//go:build marsdevfork
// +build marsdevfork

// Test-only: built only with `-tags marsdevfork` for the local smoke test.
// Never use this tag for a release binary.

package params

import "math/big"

func init() {
	MarsFastForkBlock = big.NewInt(40)
}
