# syntax=docker/dockerfile:experimental

FROM golang:1.19-alpine3.17 as build

ARG ARCH=x86_64
ARG ENABLE_WASM=false

RUN --mount=type=cache,target=/tmp/ \
  apk add --cache-dir=/tmp/ --update \
  ca-certificates \
  git \
  make \
  build-base

WORKDIR axelar

RUN --mount=type=bind,source=. \
  --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Cosmwasm - Download correct libwasmvm version
RUN if [[ "${ENABLE_WASM}" == "true" ]]; then \
    WASMVM_VERSION=v1.1.1 && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.${ARCH}.a \
        -O /lib/libwasmvm_muslc.a && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep ${ARCH} | cut -d ' ' -f 1); \
    fi

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    make ENABLE_WASM="${ENABLE_WASM}" USE_MUSLC=true build

FROM alpine:3.17

ARG USER_ID=1000
ARG GROUP_ID=1001
RUN apk add jq
COPY --from=build /go/axelar/bin/* /usr/local/bin/
RUN addgroup -S -g ${GROUP_ID} axelard && adduser -S -u ${USER_ID} axelard -G axelard
USER axelard
COPY ./entrypoint.sh /entrypoint.sh

# The home directory of axelar-core where configuration/genesis/data are stored
ENV HOME_DIR /home/axelard
# Host name for tss daemon (only necessary for validator nodes)
ENV TOFND_HOST ""
# The keyring backend type https://docs.cosmos.network/master/run-node/keyring.html
ENV AXELARD_KEYRING_BACKEND file
# The chain ID
ENV AXELARD_CHAIN_ID axelar-testnet-lisbon-3
# The file with the peer list to connect to the network
ENV PEERS_FILE ""
# Path of an existing configuration file to use (optional)
ENV CONFIG_PATH ""
# A script that runs before launching the container's process (optional)
ENV PRESTART_SCRIPT ""
# The Axelar node's moniker
ENV NODE_MONIKER ""

# Create these folders so that when they are mounted the permissions flow down
RUN mkdir /home/axelard/.axelar && chown axelard /home/axelard/.axelar
RUN mkdir /home/axelard/shared && chown axelard /home/axelard/shared
RUN mkdir /home/axelard/genesis && chown axelard /home/axelard/genesis
RUN mkdir /home/axelard/scripts && chown axelard /home/axelard/scripts
RUN mkdir /home/axelard/conf && chown axelard /home/axelard/conf

ENTRYPOINT ["/entrypoint.sh"]
