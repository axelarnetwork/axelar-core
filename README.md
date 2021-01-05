# axelar-core

The axelar-core app based on the Cosmos SDK is the main application of the axelar network.
This repository is used to build the necessary binaries and docker image to run a core validator.
See the axelarnetwork/axelarate repo for instructions to run a node. 

## Dependencies

This repository is dependent on https://github.com/axelarnetwork/tssd/. To be able to build it ensure they reside both in the same parent directory:
```
|
+--axelarnetwork
   |
   +--axelar-core
   |
   +--tssd
```
Execute `make copy-tssd` to copy the build-relevant data over before building axelar-core.

## Building binaries locally

Execute `make build` to create local binaries for the validator node. 
They are created in the `./bin` folder.

## Creating docker images
To create a regular docker image for the validator, execute `make docker-image`.
This creates the image axelar/core:latest.

To create a docker image for debugging (with [delve](https://github.com/go-delve/delve)), execute `make docker-image-debug`.
This creates the image axelar/core-debug:latest.

## Interacting with a local validator node
With a local (dockerized) validator running, the `axelarcli` binary can be used to interact with the node.
Run `./bin/axelarcli --help` after building the binaries to get information about the available commands.

## Show API documentation
Execute `GO111MODULE=off go get -u golang.org/x/tools/cmd/godoc` to ensure that `godoc` is installed on the host.

After the installation, execute `godoc -http ":{port}" -index` to host a local godoc server. For example, with port `8080` the documentation is hosted at 
http://localhost:8080/pkg/github.com/axelarnetwork/axelar-core. The index flag makes the documentation searchable.

Comments at the beginning of packages, before types and before functions are automatically taken from the source files to populate the documentation. 
See https://blog.golang.org/godoc for more information.

## Test tools
Because it is an executable, github.com/matryer/moq is not automatically downloaded when executing ``go mod download`` or similar commands. Execute ``go get github.com/matryer/moq`` to install the _moq_ tool to generate mocks for interfaces.

In [testutils](https://github.com/axelarnetwork/axelar-core/tree/master/testutils) there are helpers defined to simplify randomized testing.
