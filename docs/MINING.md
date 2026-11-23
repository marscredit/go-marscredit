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

Use these; do not invent extra enodes.

| File | What it is |
| --- | --- |
| [`genesis.json`](../genesis.json) | Chain 110110 genesis (init every node with this) |
| [`params/bootnodes.go`](../params/bootnodes.go) | Discovery bootnodes `node{1,2,3}.marscredit.xyz` |
| [`static-nodes.json`](../static-nodes.json) / [`trusted-nodes.json`](../trusted-nodes.json) | Railway static peers (same list) |
| [`configs/mainnet.env`](../configs/mainnet.env) | Same values as shell variables |

Bootnodes (also compiled into this binary):

```
enode://dbbe3272e00e5151932c852607992cb34403c87ae84536272d70d1205e1b546cc1534c6b2e9442c263647e2a5796a70ab4d726621f1b9d030c7358eb5120173e@node1.marscredit.xyz:8541
enode://234ec10bb2441e306399e6d73a62a207ca9c3a5bfdf5a149af9cc998b1444ae86869f595cf1bd4afdb35810161f5966ede85b1a1e4df41fe8194d1091ab65bd8@node2.marscredit.xyz:8542
enode://8cebc25eafc49c5dad938794914f637f9a7b8a94992072f8556ed4fb172af30caf086b77e293ea5a6424f4485e1a9dec931b90986255707433b0135dd1d1d79b@node3.marscredit.xyz:8543
```

`--bootnodes` is a comma-separated list of those three plus the four Railway enodes from `static-nodes.json`. Also copy `static-nodes.json` to `$DATADIR/geth/static-nodes.json`.

### Fork hash (`--eth.requiredblocks`)

Pin block **5145000** so you cannot follow a split:

```
--eth.requiredblocks 5145000=<full hash>
```

The canonical hash **starts with `0x622f` and ends with `6e09`**. Look it up (read-only; this is not a mining endpoint):

```shell
curl -s -X POST https://rpc.marscredit.xyz \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x4e81a8", false],"id":1}'
# or open https://blockscan.marscredit.xyz/block/5145000
```

`0x4e81a8` is 5145000. Save the `hash` field and pass it to `--eth.requiredblocks`. If it does not start with `0x622f` and end with `6e09`, you are not looking at the canonical fork block.

## Linux CPU solo mining

Full sync is required to mine (`--syncmode full`). Snap/light cannot seal.

```shell
DATADIR="$HOME/.marscredit"
ETHERBASE="0xYourAddress"          # rewards go here; no keystore needed for etherbase
THREADS="$(nproc)"                 # or a smaller number
DAGDIR="$DATADIR/ethash"
BOOTNODES="$(paste -sd, <<'EOF'
enode://dbbe3272e00e5151932c852607992cb34403c87ae84536272d70d1205e1b546cc1534c6b2e9442c263647e2a5796a70ab4d726621f1b9d030c7358eb5120173e@node1.marscredit.xyz:8541
enode://234ec10bb2441e306399e6d73a62a207ca9c3a5bfdf5a149af9cc998b1444ae86869f595cf1bd4afdb35810161f5966ede85b1a1e4df41fe8194d1091ab65bd8@node2.marscredit.xyz:8542
enode://8cebc25eafc49c5dad938794914f637f9a7b8a94992072f8556ed4fb172af30caf086b77e293ea5a6424f4485e1a9dec931b90986255707433b0135dd1d1d79b@node3.marscredit.xyz:8543
enode://bf93a274569cd009e4172c1a41b8bde1fb8d8e7cff1e5130707a0cf5be4ce0fc673c8a138ecb7705025ea4069da8c1d4b7ffc66e8666f7936aa432ce57693353@roundhouse.proxy.rlwy.net:50590
enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@monorail.proxy.rlwy.net:26254
enode://7f2ee75a1c112735aaa43de1e5a6c4d7e07d03a5352b5782ed8e0c7cc046a8c8839ad093b09649e0b4a6ed8900211fb4438765c99d07bb00006ef080a1aa9ab6@viaduct.proxy.rlwy.net:30270
enode://98710174f4798dae1931e417944ac7a7fb3268d38ef8d3941c8fcc44fe178b118003d8b3d61d85af39c561235a1708f8dd61f8ba47df4c4a6b9156e272af2cfc@monorail.proxy.rlwy.net:29138
EOF
)"

mkdir -p "$DATADIR/geth" "$DAGDIR"
cp static-nodes.json "$DATADIR/geth/static-nodes.json"

geth --datadir "$DATADIR" init genesis.json

# Replace <FORK_HASH> with the 0x622f…6e09 value from the lookup above.
geth --datadir "$DATADIR" \
  --networkid 110110 \
  --syncmode full \
  --bootnodes "$BOOTNODES" \
  --eth.requiredblocks "5145000=<FORK_HASH>" \
  --mine \
  --miner.etherbase "$ETHERBASE" \
  --miner.threads "$THREADS" \
  --ethash.dagdir "$DAGDIR" \
  --maxpeers 50 \
  --cache 2048 \
  --verbosity 3
```

No HTTP RPC is required for CPU mining. IPC at `$DATADIR/geth.ipc` is enough (`geth attach`).

CPU mining was smoke-tested on this tree: after `geth init genesis.json`, `miner.start(1)` on a throwaway etherbase logged `Updated mining threads`, `Commit new sealing work`, and `Generating DAG in progress` (epoch 0). Finding a block is not required and was not attempted. Live P2P sync against `node{1,2,3}.marscredit.xyz` / Railway TCP ports was not possible from the build environment (those hosts are blocked); run the verify-sync steps below on a machine that can reach the network.

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
  --bootnodes enode://dbbe3272e00e5151932c852607992cb34403c87ae84536272d70d1205e1b546cc1534c6b2e9442c263647e2a5796a70ab4d726621f1b9d030c7358eb5120173e@node1.marscredit.xyz:8541,enode://234ec10bb2441e306399e6d73a62a207ca9c3a5bfdf5a149af9cc998b1444ae86869f595cf1bd4afdb35810161f5966ede85b1a1e4df41fe8194d1091ab65bd8@node2.marscredit.xyz:8542,enode://8cebc25eafc49c5dad938794914f637f9a7b8a94992072f8556ed4fb172af30caf086b77e293ea5a6424f4485e1a9dec931b90986255707433b0135dd1d1d79b@node3.marscredit.xyz:8543,enode://bf93a274569cd009e4172c1a41b8bde1fb8d8e7cff1e5130707a0cf5be4ce0fc673c8a138ecb7705025ea4069da8c1d4b7ffc66e8666f7936aa432ce57693353@roundhouse.proxy.rlwy.net:50590,enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@monorail.proxy.rlwy.net:26254,enode://7f2ee75a1c112735aaa43de1e5a6c4d7e07d03a5352b5782ed8e0c7cc046a8c8839ad093b09649e0b4a6ed8900211fb4438765c99d07bb00006ef080a1aa9ab6@viaduct.proxy.rlwy.net:30270,enode://98710174f4798dae1931e417944ac7a7fb3268d38ef8d3941c8fcc44fe178b118003d8b3d61d85af39c561235a1708f8dd61f8ba47df4c4a6b9156e272af2cfc@monorail.proxy.rlwy.net:29138 \
  --eth.requiredblocks 5145000=<FORK_HASH> \
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
  --eth.requiredblocks "5145000=<FORK_HASH>" \
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

1. Install this `geth-linux-amd64` on the rig, init with `genesis.json`, copy `static-nodes.json`, and run the GPU geth command above under systemd (or a HiveOS custom service) so RPC is `http://127.0.0.1:8545`.
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

Also check https://blockscan.marscredit.xyz/block/5145000 (and any later height you pick). Same number must have the same hash. If they differ, you are on a split — stop mining and resync from genesis with `--eth.requiredblocks` set.

While syncing, `admin.peers.length` should be > 0 and `eth.syncing` should show a rising `currentBlock`. After sync, `eth.blockNumber` should track the explorer within a few blocks.

Genesis (block 0) from `genesis.json` in this repo is `0xe90d3bac8762fb4acf335b8853ca06156c8601f78a0d49ef558b7263f7a14baa`. If `eth.getBlock(0).hash` differs, you initialized the wrong file.

## Related

- Fork rules: [`MARS_FAST.md`](../MARS_FAST.md)
- Local / CI builds: [`scripts/build-linux.sh`](../scripts/build-linux.sh), [`Makefile`](../Makefile) (`linux`, `linux-amd64`, `linux-arm64`)
