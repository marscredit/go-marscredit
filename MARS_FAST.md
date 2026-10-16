# Mars Credit fast blocks (mars-fast1)

The `mars-fast-v1` branch is cut from upstream go-ethereum v1.10.18 (de23cf91). Running `geth version` prints `1.10.18-mars-fast1`.

This is a hard fork for chain 110110 at `params.MarsFastForkBlock`. The fork block is set in `params/mars.go` and keyed by chain ID, so stored chain configs don't matter. The block number is currently a **PLACEHOLDER** that has to be confirmed before release.

| | before fork | from fork block |
|---|---|---|
| difficulty | stock geth (EIP-3554 calculator, since London is at 0) | zawy12 LWMA-1, window N=120, target T=2 s, solvetimes clamped to [-6T, 6T], floor 16384, no bomb |
| first blocks | – | fork block: D = D(parent)/6. Next 10 blocks: held at that value. After that: LWMA over the post-fork blocks available, up to N |
| future-block allowance | 15 s | 2 s (`params.MarsAllowedFutureBlockTimeSeconds`) |
| fork ID | – | the fork block is included in EIP-2124 fork ID (FORK_NEXT), so stale nodes get dropped after the fork |
| reward | 2 MARS (Constantinople) | unchanged |
| miner policy | – | the local miner starts sealing a child no earlier than 1 s after it imports the parent (`miner/worker.go`, not a consensus rule) |

Code is in `params/mars.go`, `consensus/ethash/mars.go` and `consensus/ethash/consensus.go` (CalcDifficulty, verifyHeader, VerifyHeaders batch overlay, Prepare), `core/forkid/forkid.go`, and `miner/worker.go`.

Tests: run `go test ./params ./consensus/ethash ./core/forkid ./miner -run Mars`. The testdata holds 201 real mainnet headers (5,137,680 to 5,137,880). The tests check header hashes, stock difficulty and full header verification, plus an ethash light-cache PoW check on a sample of seals.

`params/mars_devfork.go` is used only with `-tags marsdevfork`. It moves the fork to block 40 for local smoke tests. Never ship a binary built with that tag.

Build: Go 1.17. Run `go build ./cmd/geth`, or `make geth`. Cross-builds use CGO_ENABLED=0 (pure-Go secp256k1).
