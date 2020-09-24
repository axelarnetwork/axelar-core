# axelar-core

The axelar-core app based on the Cosmos SDK is the main application of the axelar network.
This repository is used to build the necessary binaries and docker image to run a core validator.
See the axelarnetwork/axelarate repo for instructions to run a node. 

## Building binaries locally

Execute `make build` to create local binaries for the validator node. 
They are created in the `./bin` folder.

## Creating a docker image
To create a docker image for the validator, execute `make docker-image`.
This creates the image axelar/core:latest.

## Interacting with a local validator node
With a local (dockerized) validator running, the `axelarcli` binary can be used to interact with the node.
Run `./bin/axelarcli --help` after building the binaries to get information about the available commands.