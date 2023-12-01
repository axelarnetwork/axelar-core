# syntax=docker/dockerfile:experimental

FROM golang:1.21-alpine3.18 as build

ARG ARCH=x86_64
ARG WASM=true

RUN apk add --no-cache --update \
  ca-certificates \
  git \
  make \
  build-base \
  linux-headers

WORKDIR axelar

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# Use a compatible libwasmvm
# Alpine Linux requires static linking against muslc: https://github.com/CosmWasm/wasmd/blob/v0.33.0/INTEGRATION.md#prerequisites
RUN if [[ "${WASM}" == "true" ]]; then \
    WASMVM_VERSION=v1.3.0 && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.${ARCH}.a \
        -O /lib/libwasmvm_muslc.a && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.${ARCH}.a | cut -d ' ' -f 1); \
    fi

COPY . .

ENV CGO_ENABLED=1
RUN make MUSLC="${WASM}" build

FROM alpine:3.18

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
