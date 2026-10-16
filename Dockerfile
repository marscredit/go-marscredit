FROM golang:1.17-alpine AS build

# Install necessary packages
RUN apk add --no-cache ca-certificates make gcc musl-dev linux-headers git

# Clone the Mars Credit geth fork (mars-fast1: 2 s LWMA hard fork)
RUN git clone --branch mars-fast-v1 https://github.com/marscredit/go-marscredit.git /go-ethereum

# Set the working directory
WORKDIR /go-ethereum

# Pin the exact reviewed commit (not the moving branch head)
RUN git checkout 4d549c7fdc1af9fd10ec6ceea8cf532a0a3e050b

# Build Geth and fail the image build if it is not the fork
RUN make geth && ./build/bin/geth version | grep -q "1.10.18-mars-fast1"

# Use a minimal image for the final build
FROM alpine:latest

# Install bash
RUN apk add --no-cache bash

# Copy the Geth binary from the build stage
COPY --from=build /go-ethereum/build/bin/geth /usr/local/bin/geth

# Create necessary directories
RUN mkdir -p /data/geth/ethash && mkdir -p /data/.ethash && mkdir -p /data/geth/chaindata && mkdir -p /app/keystore

# Copy the genesis file and other necessary files
COPY genesis.json /app/genesis.json
COPY nodekey2 /app/nodekey2
COPY entrypoint_node2.sh /app/entrypoint_node2.sh
COPY static-nodes.json /data/static-nodes.json
COPY trusted-nodes.json /data/trusted-nodes.json
COPY keystore/* /app/keystore
COPY passwordfile /app/passwordfile

# Make the script executable
RUN chmod +x /app/entrypoint_node2.sh

# Create app directory
WORKDIR /app

EXPOSE 443 8545 8546 30303 30304 30305


# Use the entrypoint script
CMD ["/bin/sh", "-c", "/app/entrypoint_${NODE_ID}.sh"]