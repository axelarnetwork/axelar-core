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

RUN go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor

COPY ./go.mod .
COPY ./go.sum .
RUN --mount=type=ssh go mod download

COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:3.12

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

ENV DAEMON_HOME "$HOME_DIR/.axelar_sidecar"

ENV DAEMON_NAME "axelard"

ENV DAEMON_ALLOW_DOWNLOAD_BINARIES true

ENV DAEMON_RESTART_AFTER_UPGRADE true

COPY --from=build /go/bin/cosmovisor /usr/local/bin/cosmovisor
COPY --from=build /go/axelar/bin/axelard /usr/local/bin/axelard
COPY --from=build /go/axelar/bin/axelard $DAEMON_HOME/cosmovisor/genesis/bin/axelard
COPY ./entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
