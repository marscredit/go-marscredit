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

# Wait for private networking to initialize
sleep 10

# Start Geth and enable mining
echo "Starting Geth on node2 and enabling mining"
exec geth --datadir /data \
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
    --mine \
    --miner.threads=1 \
    --miner.etherbase 0xD21602919e81e32A456195e9cE34215Af504535A \
    --bootnodes "enode://bf93a274569cd009e4172c1a41b8bde1fb8d8e7cff1e5130707a0cf5be4ce0fc673c8a138ecb7705025ea4069da8c1d4b7ffc66e8666f7936aa432ce57693353@roundhouse.proxy.rlwy.net:50590,enode://7f2ee75a1c112735aaa43de1e5a6c4d7e07d03a5352b5782ed8e0c7cc046a8c8839ad093b09649e0b4a6ed8900211fb4438765c99d07bb00006ef080a1aa9ab6@viaduct.proxy.rlwy.net:30270,enode://98710174f4798dae1931e417944ac7a7fb3268d38ef8d3941c8fcc44fe178b118003d8b3d61d85af39c561235a1708f8dd61f8ba47df4c4a6b9156e272af2cfc@monorail.proxy.rlwy.net:29138" \
    --verbosity 6 \
    --maxpeers 50 \
    --cache 2048 \
    --nodekey /app/nodekey2 \
    --ethash.dagdir /data/.ethash &
    
# Wait indefinitely so the script doesn't exit
wait
