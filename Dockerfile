# syntax=docker/dockerfile:experimental

FROM golang:1.16-alpine3.12 as build

RUN apk add --no-cache --update \
  openssh-client \
  git \
  ca-certificates \
  make

WORKDIR axelar

RUN git config --global url."git@github.com:axelarnetwork".insteadOf https://github.com/axelarnetwork
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

COPY ./go.mod .
COPY ./go.sum .
RUN --mount=type=ssh go mod download

COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.12

COPY --from=build /go/axelar/bin/* /usr/local/bin/
COPY ./entrypoint.sh /entrypoint.sh

# The home directory of axelar-core where configuration/genesis/data are stored
ENV HOME_DIR /root
# Host name for tss daemon (only necessary for validator nodes)
ENV TOFND_HOST ""
# The keyring backend type https://docs.cosmos.network/master/run-node/keyring.html
ENV KEYRING_BACKEND test
# The chain ID
ENV CHAIN_ID axelar
# The file with the peer list to connect to the network
ENV PEERS_FILE ""
# Path of an existing configuration file to use (optional)
ENV CONFIG_PATH ""
# A initialization script to create the genesis file (optional)
ENV INIT_SCRIPT ""

ENTRYPOINT ["/entrypoint.sh"]
