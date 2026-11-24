# Mining Mars Credit (solo, your own node)

Mars Credit is an Ethash PoW L1 (chain ID **110110**). The `mars-fast1` hard fork is at block **5,145,000**: ~2 second blocks, 2 MARS reward, LWMA-1 difficulty. Miners must run **geth 1.10.18-mars-fast1**.

This document is the Linux path. The Mars Credit Miner app (macOS / Windows) ships its own pinned binary; Linux and GPU rigs should use the binaries built from this repo.

## Which binary

`geth version` must print **`1.10.18-mars-fast1`**. Anything else (stock geth, this repo's `master` without the fork, or a `-tags marsdevfork` build) will follow the wrong chain after the fork.

| Source | Pin |
| --- | --- |
| Git branch | `mars-fast-v1` |
| Geth source commit | `4d549c7fdc1af9fd10ec6ceea8cf532a0a3e050b` |
| Railway node images (`trusted-nodes/railway-node99`, `trusted-nodes/railway-node3-fast`) | Dockerfiles `git checkout 4d549c7fdc1af9fd10ec6ceea8cf532a0a3e050b` then `make geth` with Go 1.17 |

`1c09a5a` is the **ops** HEAD of `trusted-nodes/railway-node3-fast` (Dockerfile/entrypoint), not a geth source commit. That Dockerfile still builds geth from `4d549c7f`. Use `4d549c7f` for miner binaries.

**Linux binaries** (after a tagged release): GitHub Releases on this repo, `geth-linux-amd64` / `geth-linux-arm64` plus `SHA256SUMS`. Until a tag is published, build locally:

```shell
# Go 1.17.x recommended (CI uses 1.17.13). CGO is off for cross-builds.
git checkout 4d549c7fdc1af9fd10ec6ceea8cf532a0a3e050b   # or this branch after merge
make linux-amd64          # -> dist/geth-linux-amd64
# or both arches + checksums:
make linux                # -> dist/geth-linux-{amd64,arm64} and dist/SHA256SUMS
install -m 0755 dist/geth-linux-amd64 /usr/local/bin/geth
geth version              # must contain 1.10.18-mars-fast1
```

The GitHub Actions workflow `.github/workflows/release-linux.yml` builds both arches and attaches them to a **draft** GitHub Release when you push a tag matching `v*` (for example `v1.10.18-mars-fast1`).

## Hard rules

- Solo mine on **your own node**. No mining pool. No public stratum.
- **Never** point a miner at `https://rpc.marscredit.xyz` or at node99. Those are for wallet/read RPC and explorers, not for `eth_getWork` / sealing.
- **Never** expose the miner API (`miner` JSON-RPC) on a public interface. Keep HTTP RPC on `127.0.0.1` and enable only the namespaces you need.
- `--http.addr 0.0.0.0`, `--http.api …miner…`, and `--http.vhosts *` appear in some Railway entrypoints for internal ops. Do not copy that onto a miner.

## Network files in this repo

`genesis.json` and `static-nodes.json` live in this repository. Download them from the branch (current peers) or the release tag (may lag until the next tag):

| File | Branch (`mars-fast-v1`) | Tag (`v1.10.18-mars-fast1`) |
| --- | --- | --- |
| genesis | https://raw.githubusercontent.com/marscredit/go-marscredit/mars-fast-v1/genesis.json | https://raw.githubusercontent.com/marscredit/go-marscredit/v1.10.18-mars-fast1/genesis.json |
| static-nodes | https://raw.githubusercontent.com/marscredit/go-marscredit/mars-fast-v1/static-nodes.json | https://raw.githubusercontent.com/marscredit/go-marscredit/v1.10.18-mars-fast1/static-nodes.json |

Use these; do not invent extra enodes.

| File | What it is |
| --- | --- |
| [`genesis.json`](../genesis.json) | Chain 110110 genesis (init every node with this) |
| [`params/bootnodes.go`](../params/bootnodes.go) | Compiled-in discovery peers (the two live Railway enodes below) |
| [`static-nodes.json`](../static-nodes.json) / [`trusted-nodes.json`](../trusted-nodes.json) | Same two live peers |
| [`configs/mainnet.env`](../configs/mainnet.env) | Same values as shell variables |

Live public peers (checked 2026-11-23; same pair as https://marscredit.xyz/start-mining):

```
enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@roundhouse.proxy.rlwy.net:26254
enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@interchange.proxy.rlwy.net:42747
```

`--bootnodes` is a comma-separated list of those two. Also copy `static-nodes.json` to `$DATADIR/geth/static-nodes.json`.

Not live (do not dial): `node1.marscredit.xyz` / `node2.marscredit.xyz` / `node3.marscredit.xyz` (no DNS); `bf93a274…@roundhouse:50590`; `ca363906…@monorail:26254` (wrong host); `7f2ee75a…@viaduct:30270`; `98710174…@monorail:29138`.

### Fork hash (`--eth.requiredblocks`)

Every running node must pin block **5145000** (and `--networkid 110110`) so it cannot follow a split:

```
--networkid 110110
--eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09
```

That hash starts with `0x622f` and ends with `6e09`. Confirm it still matches (read-only; this is not a mining endpoint):

```shell
curl -s -X POST https://rpc.marscredit.xyz \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x4e81a8", false],"id":1}'
# or open https://blockscan.marscredit.xyz/block/5145000
```

`0x4e81a8` is 5145000. If the public `hash` is not `0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09`, stop and do not mine.

## Linux CPU solo mining

Full sync is required to mine (`--syncmode full`). Snap/light cannot seal.

```shell
DATADIR="$HOME/.marscredit"
ETHERBASE="0xYourAddress"          # rewards go here; no keystore needed for etherbase
THREADS="$(nproc)"                 # or a smaller number
DAGDIR="$DATADIR/ethash"
BOOTNODES="$(paste -sd, <<'EOF'
enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@roundhouse.proxy.rlwy.net:26254
enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@interchange.proxy.rlwy.net:42747
EOF
)"

mkdir -p "$DATADIR/geth" "$DAGDIR"
cp static-nodes.json "$DATADIR/geth/static-nodes.json"

geth --datadir "$DATADIR" init genesis.json

geth --datadir "$DATADIR" \
  --networkid 110110 \
  --syncmode full \
  --bootnodes "$BOOTNODES" \
  --eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09 \
  --mine \
  --miner.etherbase "$ETHERBASE" \
  --miner.threads "$THREADS" \
  --ethash.dagdir "$DAGDIR" \
  --maxpeers 50 \
  --cache 2048 \
  --verbosity 3
```

No HTTP RPC is required for CPU mining. IPC at `$DATADIR/geth.ipc` is enough (`geth attach`).

Check peers:

```shell
geth --exec 'net.peerCount' attach "$DATADIR/geth.ipc"
```

If `net.peerCount` is 0, add the two live peers by hand (fallback if `--bootnodes` / `static-nodes.json` did not stick):

```shell
geth --exec 'admin.addPeer("enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@roundhouse.proxy.rlwy.net:26254")' attach "$DATADIR/geth.ipc"
geth --exec 'admin.addPeer("enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@interchange.proxy.rlwy.net:42747")' attach "$DATADIR/geth.ipc"
geth --exec 'net.peerCount' attach "$DATADIR/geth.ipc"
```

You want `net.peerCount` ≥ 1 before you care about hashrate. `admin.addPeer` is over IPC (not HTTP).

CPU mining was smoke-tested on this tree: after `geth init genesis.json`, `miner.start(1)` on a throwaway etherbase logged `Updated mining threads`, `Commit new sealing work`, and `Generating DAG in progress` (epoch 0). Finding a block is not required and was not attempted. Live P2P from the original build environment was blocked by egress; the two enodes above were confirmed from a Mac with working P2P on 2026-11-23.

### systemd (`Restart=on-failure`)

`/etc/systemd/system/geth-marscredit.service`:

```ini
[Unit]
Description=Mars Credit geth (1.10.18-mars-fast1)
After=network-online.target
Wants=network-online.target

[Service]
User=marscredit
Group=marscredit
Type=simple
Restart=on-failure
RestartSec=5
TimeoutStopSec=90
LimitNOFILE=65535
Environment=HOME=/var/lib/marscredit
ExecStart=/usr/local/bin/geth \
  --datadir /var/lib/marscredit \
  --networkid 110110 \
  --syncmode full \
  --bootnodes enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@roundhouse.proxy.rlwy.net:26254,enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@interchange.proxy.rlwy.net:42747 \
  --eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09 \
  --mine \
  --miner.etherbase 0xYourAddress \
  --miner.threads 4 \
  --ethash.dagdir /var/lib/marscredit/ethash \
  --maxpeers 50 \
  --cache 2048 \
  --verbosity 3

[Install]
WantedBy=multi-user.target
```

```shell
sudo install -d -o marscredit -g marscredit /var/lib/marscredit/geth /var/lib/marscredit/ethash
sudo -u marscredit geth --datadir /var/lib/marscredit init /opt/go-marscredit/genesis.json
sudo -u marscredit cp /opt/go-marscredit/static-nodes.json /var/lib/marscredit/geth/static-nodes.json
sudo systemctl daemon-reload
sudo systemctl enable --now geth-marscredit
```

This unit file was not executed on a real systemd host here; flags match the CPU command above.

## GPU solo mining (local node + ethminer)

Run geth on the same machine (or another host on your LAN that is **not** publicly reachable). Bind HTTP RPC to localhost only. Enable `eth` (getWork lives there). Do **not** enable `miner`, `personal`, or `admin` on HTTP.

`--miner.threads 0` turns off CPU sealing so the GPU miner is the only hasher. `--mine` still has to be on so geth produces work.

```shell
geth --datadir "$DATADIR" \
  --networkid 110110 \
  --syncmode full \
  --bootnodes "$BOOTNODES" \
  --eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09 \
  --http --http.addr 127.0.0.1 --http.port 8545 --http.api eth,net,web3 \
  --mine --miner.threads 0 --miner.etherbase "$ETHERBASE" \
  --ethash.dagdir "$DAGDIR" \
  --maxpeers 50 --cache 2048 --verbosity 3
```

Then, on the same machine:

```shell
ethminer -G -P http://127.0.0.1:8545
# CUDA:  ethminer -U -P http://127.0.0.1:8545
# OpenCL: ethminer -G -P http://127.0.0.1:8545
```

**Untested here** (no GPU in the build environment). Flags are the stock geth 1.10.18 getWork path. If ethminer cannot fetch work, confirm `eth.syncing === false` and `eth.mining === true` over IPC, and that nothing else is bound to 8545.

Do not pass `--http.addr 0.0.0.0` “so HiveOS on another box can connect” unless that other box is on a private network you control. Prefer running geth on the rig itself.

### HiveOS custom miner (local node)

**Untested** on HiveOS. Outline only:

1. Install this `geth-linux-amd64` on the rig. Download `genesis.json` and `static-nodes.json` from the raw GitHub URLs above, init, copy static-nodes into `$DATADIR/geth/`, and run geth locally (systemd or a HiveOS custom service) with RPC on `http://127.0.0.1:8545` and the two live `--bootnodes`. After start, `geth --exec 'net.peerCount' attach "$DATADIR/geth.ipc"` should be ≥ 1; if not, use the `admin.addPeer` fallback.

```shell
geth --datadir "$DATADIR" \
  --networkid 110110 \
  --syncmode full \
  --bootnodes "$BOOTNODES" \
  --eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09 \
  --http --http.addr 127.0.0.1 --http.port 8545 --http.api eth,net,web3 \
  --mine --miner.threads 0 --miner.etherbase "$ETHERBASE" \
  --ethash.dagdir "$DAGDIR" \
  --maxpeers 50 --cache 2048 --verbosity 3
```

2. HiveOS → miner → **custom miner**. Point the miner URL at `http://127.0.0.1:8545` (getWork), not a stratum `stratum+tcp://…` and not `rpc.marscredit.xyz`.
3. Wallet / etherbase is geth's `--miner.etherbase`, not a pool username.

If a HiveOS template asks for a pool host, you are in the wrong UI. There is no official Mars Credit pool or public stratum.

## Verify you are on the canonical chain

Do this **before** you care about hashrate. Compare a block you have locally with the public explorer/RPC (read-only):

```shell
# Local (IPC). Use a height you have already imported.
geth --exec 'eth.getBlock(5145000).hash' attach "$DATADIR/geth.ipc"

# Public (lookup only — not for mining):
curl -s -X POST https://rpc.marscredit.xyz \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x4e81a8", false],"id":1}'
```

Also check https://blockscan.marscredit.xyz/block/5145000 (and any later height you pick). Block 5145000 must be `0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09`. If they differ, you are on a split — stop mining and resync from genesis with `--networkid 110110 --eth.requiredblocks 5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09`.

While syncing, `net.peerCount` should be > 0 and `eth.syncing` should show a rising `currentBlock`. If `net.peerCount` stays 0, run the `admin.addPeer` fallback above. After sync, `eth.blockNumber` should track the explorer within a few blocks.

Genesis (block 0) from `genesis.json` in this repo is `0xe90d3bac8762fb4acf335b8853ca06156c8601f78a0d49ef558b7263f7a14baa`. If `eth.getBlock(0).hash` differs, you initialized the wrong file. Fetch genesis and static-nodes from the raw GitHub URLs in [Network files](#network-files-in-this-repo).

## Related

- Fork rules: [`MARS_FAST.md`](../MARS_FAST.md)
- Local / CI builds: [`scripts/build-linux.sh`](../scripts/build-linux.sh), [`Makefile`](../Makefile) (`linux`, `linux-amd64`, `linux-arm64`)
