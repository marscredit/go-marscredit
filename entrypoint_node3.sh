#!/bin/sh

echo "Starting Node 3 (node3-fast, mars-fast1)"

NODE99_ENODE="enode://478d4dd8e56a6419d32a7f1166983566bf5c184d1fd97036bca04a2f8e5a6e7546492e38ebf16fbdb0c838c9d127ee70b5de6a7e1cdc5d337adc616f91baf98f@node99.railway.internal:30304"
NODE2_ENODE="enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@node2.railway.internal:30304"

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

# Generate nodekey if not present
echo "Checking for nodekey..."
if [ ! -f "/app/nodekey3" ]; then
    echo "Generating new nodekey..."
    bootnode -genkey /app/nodekey3
else
    echo "nodekey3 file exists."
fi

# Initialize Geth with the genesis file (only needed for first run)
if [ -z "$(ls -A /data/geth/chaindata)" ]; then
    echo "Chaindata directory is empty. Initializing Geth with genesis file."
    geth init /app/genesis.json --datadir /data
else
    echo "Chaindata directory exists and is not empty."
fi

# Static peers: node99 and node2 over Railway private networking.
# (The image's /data/static-nodes.json is hidden by the volume mount.)
cat > /data/geth/static-nodes.json <<JSON
[
  "$NODE99_ENODE",
  "$NODE2_ENODE"
]
JSON
echo "Wrote /data/geth/static-nodes.json"

# Wait for private networking to initialize
sleep 10

# Start Geth (mining ON, rewards to 0xD216...535A; public HTTP/WS limited to eth,net,web3; no account unlock)
# Note: geth only starts sealing once it has finished syncing.
echo "Starting Geth on node3 (mining enabled)"
geth --datadir /data \
    --syncmode "full" \
    --http \
    --http.addr "0.0.0.0" \
    --http.port 8546 \
    --http.api eth,net,web3 \
    --http.vhosts=* \
    --http.corsdomain=* \
    --networkid 110110 \
    --ws \
    --ws.addr "0.0.0.0" \
    --ws.port 8546 \
    --ws.api eth,net,web3 \
    --port 30304 \
    --nat "any" \
    --bootnodes "$NODE99_ENODE,$NODE2_ENODE" \
    --verbosity 3 \
    --maxpeers 50 \
    --cache 2048 \
    --nodekey /app/nodekey3 \
    --ethash.dagdir /data/.ethash \
    --mine \
    --miner.etherbase 0xD21602919e81e32A456195e9cE34215Af504535A \
    --miner.threads 8 &
GETH_PID=$!

# Wait for geth; the TERM/INT trap waits for geth to exit cleanly
wait "$GETH_PID"
