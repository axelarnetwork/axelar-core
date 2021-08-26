#!/usr/bin/env bash

container_id=$(docker create axelar/core:binaries)
docker cp "$container_id":/go/src/github.com/axelarnetwork/axelar-core/bin ./
docker rm -v "$container_id"
