# syntax=docker/dockerfile:latest

####################################
### build base
####################################
FROM golang:1.18-alpine3.16 as build-base

RUN --mount=type=cache,target=/tmp/ \
  apk add --cache-dir=/tmp/ --update \
  ca-certificates \
  git \
  make

WORKDIR axelar

RUN --mount=type=bind,source=. \
  --mount=type=cache,target=/go/pkg/mod \
  go mod download

COPY . .
ENV CGO_ENABLED=0

####################################
### prod base
####################################
FROM alpine:3.16 as prod-base

# The home directory of axelar-core where configuration/genesis/data are stored
ENV HOME_DIR=/home/axelard
# Host name for tss daemon (only necessary for validator nodes)
ENV TOFND_HOST=""
# The keyring backend type https://docs.cosmos.network/master/run-node/keyring.html
ENV AXELARD_KEYRING_BACKEND=file
# The chain ID
ENV AXELARD_CHAIN_ID=axelar-testnet-lisbon-3
# The file with the peer list to connect to the network
ENV PEERS_FILE=""
# Path of an existing configuration file to use (optional)
ENV CONFIG_PATH=""
# A script that runs before launching the container's process (optional)
ENV PRESTART_SCRIPT=""
# The Axelar node's moniker
ENV NODE_MONIKER=""

ARG USER_ID=1000
ARG GROUP_ID=1001
RUN addgroup -S -g ${GROUP_ID} axelard; \
  adduser -S -u ${USER_ID} axelard -G axelard; \
  # Create these folders so that when they are mounted the permissions flow down \
  apk add --no-cache jq; \
  mkdir \
    ${HOME_DIR}/.axelar \
    ${HOME_DIR}/shared \
    ${HOME_DIR}/genesis \
    ${HOME_DIR}/scripts \
    ${HOME_DIR}/conf; \
  chown -cR axelard \
    ${HOME_DIR}/.axelar \
    ${HOME_DIR}/shared \
    ${HOME_DIR}/genesis \
    ${HOME_DIR}/scripts \
    ${HOME_DIR}/conf
USER axelard
ENTRYPOINT ["/entrypoint.sh"]

####################################
### make build
####################################
FROM build-base as build
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/root/.go-build \
  make build

####################################
### make build-binaries (Dockerfile.binaries)
####################################
FROM build-base as build-binaries
ARG SEMVER
ENV SEMVER=${SEMVER}
RUN apk add --no-cache bash
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/root/.go-build \
  make build-binaries

####################################
### debug image (Dockerfile.debug)
####################################
FROM build-base as delve
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/root/.go-build \
  go install github.com/go-delve/delve/cmd/dlv@latest

FROM prod-base as debug
# Whether or not to start the REST server
ENV START_REST=false
# The keyring backend type https://docs.cosmos.network/master/run-node/keyring.html
ENV KEYRING_BACKEND=test
# Should dlv wait for a debugger to attach to the rest server process before starting it?
ENV REST_CONTINUE=true
# Should dlv wait for a debugger to attach to the axelard process before starting it?
ENV CORE_CONTINUE=true
# Debug mode
ENV DEBUG_MODE=true
COPY --from=build /go/axelar/entrypoint.sh /entrypoint.sh
COPY --from=delve /go/bin/dlv /usr/local/bin/dlv
COPY --from=build /go/axelar/bin/* /usr/local/bin/

ENTRYPOINT ["/entrypoint.sh"]

####################################
### rosetta (Dockerfile.rosetta)
####################################
FROM build-base as rosetta-build
COPY rosetta/go.mod rosetta/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/root/.go-build \
  make build

FROM prod-base as rosetta
COPY ./rosetta/entrypoint.sh /entrypoint.sh
COPY --from=rosetta-build /go/axelar/bin/* /usr/local/bin/

####################################
### default build prod image on docker build
####################################
FROM prod-base
COPY --from=build /go/axelar/entrypoint.sh /entrypoint.sh
COPY --from=build /go/axelar/bin/* /usr/local/bin/
