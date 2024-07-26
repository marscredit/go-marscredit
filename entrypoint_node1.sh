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
if [ -z "$(ls -A /data/geth/chaindata)" ]; then
    echo "Chaindata directory is empty. Initializing Geth with genesis file."
    geth init /app/genesis.json --datadir /data
else
    echo "Chaindata directory exists and is not empty."
fi

# echo "Chaindata directory is empty. Initializing Geth with genesis file."
# geth init /app/genesis.json --datadir /data

# Start Geth and enable mining
echo "Starting Geth and enabling mining"
geth --datadir /data \
    --keystore /app/keystore \
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
    --bootnodes="enode://ca3639067a580a0f1db7412aeeef6d5d5e93606ed7f236a5343fe0d1115fb8c2bea2a22fa86e9794b544f886a4cb0de1afcbccf60960802bf00d81dab9553ec9@monorail.proxy.rlwy.net:26254,enode://7f2ee75a1c112735aaa43de1e5a6c4d7e07d03a5352b5782ed8e0c7cc046a8c8839ad093b09649e0b4a6ed8900211fb4438765c99d07bb00006ef080a1aa9ab6@viaduct.proxy.rlwy.net:30270,enode://98710174f4798dae1931e417944ac7a7fb3268d38ef8d3941c8fcc44fe178b118003d8b3d61d85af39c561235a1708f8dd61f8ba47df4c4a6b9156e272af2cfc@monorail.proxy.rlwy.net:29138" \
    --verbosity 6 \
    --maxpeers 50 \
    --cache 2048 \
    --nodiscover \
    --nodekey /data/geth/nodekey \
    --ethash.dagdir /data/.ethash &

# Wait indefinitely so the script doesn't exit
wait
