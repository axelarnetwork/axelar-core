# syntax=docker/dockerfile:experimental

FROM golang:1.18-bullseye as build

RUN apt update && apt install ca-certificates git make -y

WORKDIR axelar

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM ubuntu:20.04
RUN apt update && apt install jq -y
ARG USER_ID=1000
ARG GROUP_ID=1001
COPY --from=build /go/axelar/bin/* /usr/local/bin/
RUN useradd --uid ${USER_ID} axelard && groupmod --gid ${GROUP_ID} axelard && usermod -aG axelard axelard
RUN chown axelard /home && mv bin/axelard /usr/local/bin && chmod 777 /usr/local/bin/axelard
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
RUN mkdir -p /home/axelard/.axelar && chown axelard /home/axelard/.axelar
RUN mkdir -p /home/axelard/shared && chown axelard /home/axelard/shared
RUN mkdir -p /home/axelard/genesis && chown axelard /home/axelard/genesis
RUN mkdir -p /home/axelard/scripts && chown axelard /home/axelard/scripts
RUN mkdir -p /home/axelard/conf && chown axelard /home/axelard/conf


ENTRYPOINT ["/entrypoint.sh"]
