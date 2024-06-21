#!/bin/sh

echo "Starting Node 1"

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
mkdir -p /app/keystore

# Set permissions to ensure Geth can write to the directory
chmod -R 755 /app
chmod -R 755 /data

# Log the contents of the directories
echo "---- Logging contents of /app:"
ls -la /app

echo "---- Logging contents of /data:"
ls -la /data

echo "---- Logging contents of /app/keystore:"
ls -la /app/keystore

echo "---- Checking for specific key file:"
KEY_FILE="/app/keystore/UTC--2024-06-12T21-51-26.975004000Z--c1133a2b8e92a747ebf2a937be3d79c29231f407"
if [ -f "$KEY_FILE" ]; then
    echo "Key file $KEY_FILE exists."
else
    echo "Key file $KEY_FILE does not exist."
    exit 1
fi

# Explicitly log and check permissions of the keystore file
ls -la $KEY_FILE
cat $KEY_FILE

echo "Logging contents of /data/geth/chaindata (if exists):"
if [ -d /data/geth/chaindata ]; then
    ls -la /data/geth/chaindata
else
    echo "chaindata directory does not exist, creating now"
    mkdir -p /data/geth/chaindata
fi

# Initialize Geth with the genesis file (only needed for first run)
# if [ ! -f "/data/geth/chaindata/CURRENT" ]; then
#     echo "Initializing Geth with genesis file"
#     geth init /app/genesis.json --datadir /data
# else
#     echo "chaindata directory exists and is not empty"
# fi

geth init /app/genesis.json --datadir /data

# Start Geth and enable mining
echo "Starting Geth and enabling mining"
geth --datadir /data \
    --keystore /app/keystore \
    --syncmode "full" \
    --http \
    --http.addr "0.0.0.0" \
    --http.port 30303 \
    --http.api personal,eth,net,web3,miner,admin \
    --http.vhosts=* \
    --http.corsdomain=* \
    --networkid 110110 \
    --ws \
    --ws.addr "0.0.0.0" \
    --ws.port 30303 \
    --port 443 \
    --mine \
    --miner.threads=1 \
    --miner.etherbase 0xc1133A2B8E92a747eBF2A937bE3D79c29231f407 \
    --unlock 0xc1133A2B8E92a747eBF2A937bE3D79c29231f407 \
    --password /app/passwordfile \
    --allow-insecure-unlock \
    --verbosity 6 \
    --maxpeers 50 \
    --cache 2048 \
    --nodiscover \
    --nodekey /data/geth/nodekey \
    --ethash.dagdir /data/.ethash &

# Add a loop to keep the script running
while true; do
  sleep 60
done
