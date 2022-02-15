⚠️⚠️⚠️ **THIS IS A WORK IN PROGRESS** ⚠️⚠️⚠️

# axelar-core

The axelar-core app based on the Cosmos SDK is the main application of the axelar network. This repository is used to
build the necessary binaries and docker image to run a core node.

## Building binaries locally

Execute `make build` to create local binaries for the validator node. They are created in the `./bin` folder.

## Creating docker images

To create a regular docker image for the node, execute `make docker-image`. This creates the image axelar/core:
latest.

To create a docker image for debugging (with [delve](https://github.com/go-delve/delve)),
execute `make docker-image-debug`. This creates the image axelar/core-debug:latest.

## Interacting with a local node

With a local (dockerized) node running, the `axelard` binary can be used to interact with the node.
Run `./bin/axelard --help` after building the binaries to get information about the available commands.

## Show API documentation

Execute `GO111MODULE=off go get -u golang.org/x/tools/cmd/godoc` to ensure that `godoc` is installed on the host.

After the installation, execute `godoc -http ":{port}" -index` to host a local godoc server. For example, with
port `8080` the documentation is hosted at
http://localhost:8080/pkg/github.com/axelarnetwork/axelar-core. The index flag makes the documentation searchable.

Comments at the beginning of packages, before types and before functions are automatically taken from the source files
to populate the documentation. See https://blog.golang.org/godoc for more information.

### CLI command documentation

For the full list of available CLI commands for `axelard` see [here](docs/cli/toc.md)

## Test tools

Because it is an executable, github.com/matryer/moq is not automatically downloaded when executing ``go mod download``
or similar commands. Execute ``go get github.com/matryer/moq`` to install the _moq_ tool to generate mocks for
interfaces.

## Bug bounty and disclosure of vulnerabilities

See the [Axelar documentation website](https://docs.axelar.dev/#/bug-bounty).
