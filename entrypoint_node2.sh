#!/bin/sh

echo "Starting Node 2"

# Function to handle shutdown
shutdown() {
    echo "Shutting down Geth gracefully..."
    kill -TERM "$GETH_PID" 2>/dev/null
    wait "$GETH_PID"
    echo "Geth has been shut down."
    exit 0
}

# Trap SIGTERM/SIGINT (sent by Railway when stopping the container)
trap shutdown TERM INT

mkdir -p /app/geth/ethash
mkdir -p /data/.ethash
mkdir -p /data/geth/chaindata

# Set permissions to ensure Geth can write to the directory
chmod -R 755 /app
chmod -R 755 /data

# Log the contents of the directories
echo "---- Logging contents of /app:"
ls -la /app

echo "---- Logging contents of /data:"
ls -la /data

echo "Logging contents of /data/geth/chaindata (if exists):"
if [ -d /data/geth/chaindata ]; then
    ls -la /data/geth/chaindata
else
    echo "chaindata directory does not exist, creating now"
    mkdir -p /data/geth/chaindata
fi

# Generate nodekey if not present
echo "Checking for nodekey..."
if [ ! -f "/app/nodekey2" ]; then
    echo "Generating new nodekey..."
    bootnode -genkey /app/nodekey2
else
    echo "nodekey2 file exists."
fi

# Log the contents of the nodekey
echo "Logging contents of /app/nodekey2 (if exists):"
if [ -f /app/nodekey2 ]; then
    ls -la /app/nodekey2
fi

# Initialize Geth with the genesis file (only needed for first run)
if [ -z "$(ls -A /data/geth/chaindata)" ]; then
    echo "Chaindata directory is empty. Initializing Geth with genesis file."
    geth init /app/genesis.json --datadir /data
else
    echo "Chaindata directory exists and is not empty."
fi

# Peers: the other two Railway nodes over private networking (node3-fast, NOT
# node3, which is the old stock-geth service). Written into the datadir because
# the image's /data copies are hidden by the volume mount. Trusted = kept even
# when full.
NODE99_ENODE="enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@node99.railway.internal:30304"
NODE3_FAST_ENODE="enode://7f2ee75a1c112735aaa43de1e5a6c4d7e07d03a5352b5782ed8e0c7cc046a8c8839ad093b09649e0b4a6ed8900211fb4438765c99d07bb00006ef080a1aa9ab6@node3-fast.railway.internal:30304"
cat > /data/geth/static-nodes.json <<JSON
[
  "$NODE99_ENODE",
  "$NODE3_FAST_ENODE"
]
JSON
cp /data/geth/static-nodes.json /data/geth/trusted-nodes.json
echo "Wrote /data/geth/static-nodes.json and trusted-nodes.json"

# Wait for private networking to initialize
sleep 10

# Start Geth (mining ON, rewards to 0xD216...535A; peers with node99 so it mines the shared chain)
echo "Starting Geth on node2 (mining enabled)"
geth --datadir /data \
    --syncmode "full" \
    --http \
    --http.addr "0.0.0.0" \
    --http.port 8546 \
    --http.api personal,eth,net,web3,miner,admin \
    --http.vhosts=* \
    --http.corsdomain=* \
    --networkid 110110 \
    --ws \
    --ws.addr "0.0.0.0" \
    --ws.port 8546 \
    --port 30304 \
    --nat "any" \
    --bootnodes "$NODE99_ENODE,$NODE3_FAST_ENODE" \
    --verbosity 3 \
    --eth.requiredblocks "5145000=0x622f1c0928884b81e631d50b69f44c58fd071852959c09c221c021ddd6571e09" \
    --maxpeers 50 \
    --cache 512 \
    --nodekey /app/nodekey2 \
    --ethash.dagdir /data/.ethash \
    --mine \
    --miner.etherbase 0xD21602919e81e32A456195e9cE34215Af504535A \
    --miner.threads 4 &
GETH_PID=$!

# Wait for geth; the SIGTERM trap waits for geth to exit cleanly
wait "$GETH_PID"