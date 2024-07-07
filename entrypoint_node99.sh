#!/bin/sh

echo "Starting Node 2"

# Function to handle shutdown
shutdown() {
    echo "Shutting down Geth gracefully..."
    pkill -SIGTERM geth
    wait $!
    echo "Geth has been shut down."
    exit 0
}

# Trap SIGTERM signal (sent by Railway when stopping the container)
trap shutdown SIGTERM

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
if [ ! -f "/app/nodekey99" ]; then
    echo "Generating new nodekey..."
    bootnode -genkey /app/nodekey99
else
    echo "nodekey99 file exists."
fi

# Log the contents of the nodekey
echo "Logging contents of /app/nodekey99 (if exists):"
if [ -f /app/nodekey99 ]; then
    ls -la /app/nodekey99
fi

# Initialize Geth with the genesis file (only needed for first run)
if [ -z "$(ls -A /data/geth/chaindata)" ]; then
    echo "Chaindata directory is empty. Initializing Geth with genesis file."
    geth init /app/genesis.json --datadir /data
else
    echo "Chaindata directory exists and is not empty."
fi

# Wait for private networking to initialize
sleep 10

# Start Geth and enable mining
echo "Starting Geth on node99 and enabling mining"
exec geth --datadir /data \
    --syncmode "full" \
    --http \
    --http.addr 0.0.0.0 \
    --http.port 8546 \
    --http.api personal,eth,net,web3,miner \
    --http.vhosts=* \
    --http.corsdomain=* \
    --networkid 110110 \
    --ws \
    --ws.addr 0.0.0.0 \
    --ws.port 8546 \
    --port 30304 \
    --nat "auto" \
    --mine \
    --miner.threads=1 \
    --miner.etherbase 0xD21602919e81e32A456195e9cE34215Af504535A \
    --bootnodes "enode://bf93a274569cd009e4172c1a41b8bde1fb8d8e7cff1e5130707a0cf5be4ce0fc673c8a138ecb7705025ea4069da8c1d4b7ffc66e8666f7936aa432ce57693353@roundhouse.proxy.rlwy.net:50590" \
    --verbosity 6 \
    --maxpeers 100 \
    --cache 2048 \
    --nodiscover \
    --nodekey /app/nodekey99 \
    --ethash.dagdir /data/.ethash &
    
# Wait indefinitely so the script doesn't exit
wait