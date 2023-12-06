PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
MAX_WASM_SIZE := $(shell echo "$$((3 * 1024 * 1024))")

DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
HTTPS_GIT := https://github.com/axelarnetwork/axelar-core.git
PUSH_DOCKER_IMAGE := true

$(info $$WASM is [${WASM}])
$(info $$CGO_ENABLED is [${CGO_ENABLED}])

ifeq ($(WASM), true)
WASM_ENABLED := "true"
ifndef $(WASM_CAPABILITIES)
# Wasm capabilities: https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
WASM_CAPABILITIES := "iterator,staking,stargate,cosmwasm_1_3"
endif
else
WASM_ENABLED := ""
WASM_CAPABILITIES := ""
endif

ifeq ($(MUSLC), true)
STATIC_LINK_FLAGS := -linkmode=external -extldflags '-Wl,-z,muldefs -static'
BUILD_TAGS := ledger,muslc
else
STATIC_LINK_FLAGS := ""
BUILD_TAGS := ledger
endif

ARCH := x86_64
ifeq ($(shell uname -m), arm64)
ARCH := aarch64
endif

DENOM := uaxl

ldflags = "-X github.com/cosmos/cosmos-sdk/version.Name=axelar \
	-X github.com/cosmos/cosmos-sdk/version.AppName=axelard \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(BUILD_TAGS)" \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/CosmWasm/wasmd/x/wasm/types/MaxWasmSize=${MAX_WASM_SIZE} \
	-X github.com/axelarnetwork/axelar-core/x/axelarnet/exported.NativeAsset=$(DENOM) \
	-X github.com/axelarnetwork/axelar-core/app.WasmEnabled=$(WASM_ENABLED) \
	-X github.com/axelarnetwork/axelar-core/app.WasmCapabilities=$(WASM_CAPABILITIES) \
	-w -s ${STATIC_LINK_FLAGS}"

BUILD_FLAGS := -tags $(BUILD_TAGS) -ldflags $(ldflags) -trimpath
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
OS := $(shell echo $$OS_TYPE | sed -e 's/ubuntu-20.04/linux/; s/macos-latest/darwin/')
SUFFIX := $(shell echo $$PLATFORM | sed 's/\//-/' | sed 's/\///')

.PHONY: all
all: generate goimports lint build docker-image docker-image-debug

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

# Uncomment when you have some tests
# test:
# 	@go test -mod=readonly $(PACKAGES)
.PHONY: lint
# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run
	@go mod verify

.PHONY: goimports
goimports:
	@echo "running goimports"
# exclude mocks, statik and proto generated files
	@./scripts/rm-blank-lines.sh # remove blank lines from imports
	@goimports -l -local github.com/axelarnetwork/ . | grep -v .pb.go$ | grep -v .pb.gw.go$ | grep -v mock | grep -v statik.go$ | xargs goimports -local github.com/axelarnetwork/ -w

# Build the project with release flags
.PHONY: build
build: go.sum
		go build -o ./bin/axelard -mod=readonly $(BUILD_FLAGS) ./cmd/axelard

.PHONY: build-binaries
build-binaries:  guard-SEMVER
	./scripts/build-binaries.sh ${SEMVER} '$(BUILD_TAGS)' '$(ldflags)'

# Build the project with release flags for multiarch
.PHONY: build-binaries-multiarch
build-binaries-multiarch: go.sum
		GOOS=${OS} GOARCH=${ARCH} go build -o ./bin/axelard -mod=readonly $(BUILD_FLAGS) ./cmd/axelard

.PHONY: build-binaries-in-docker
build-binaries-in-docker:  guard-SEMVER
	DOCKER_BUILDKIT=1 docker build \
		--build-arg SEMVER=${SEMVER} \
		-t axelar/core:binaries \
		-f Dockerfile.binaries .
	./scripts/copy-binaries-from-image.sh

# Build the project with debug flags
.PHONY: debug
debug:  go.sum
		go build -o ./bin/axelard -mod=readonly $(BUILD_FLAGS) -gcflags="all=-N -l" ./cmd/axelard

# Build a release image
.PHONY: docker-image
docker-image:
	@DOCKER_BUILDKIT=1 docker build \
		--build-arg WASM="${WASM}" \
		--build-arg ARCH="${ARCH}" \
		-t axelar/core .

# Build a release image
.PHONY: docker-image-local-user
docker-image-local-user:  guard-VERSION guard-GROUP_ID guard-USER_ID
	@DOCKER_BUILDKIT=1 docker build \
		--build-arg USER_ID=${USER_ID} \
		--build-arg GROUP_ID=${GROUP_ID} \
		--build-arg WASM="${WASM}" \
		--build-arg ARCH="${ARCH}" \
		-t axelarnet/axelar-core:${VERSION}-local .

.PHONY: build-push-docker-image
build-push-docker-images:  guard-SEMVER
	@DOCKER_BUILDKIT=1 docker buildx build \
		--platform ${PLATFORM} \
		--output "type=image,push=${PUSH_DOCKER_IMAGE}" \
		--build-arg WASM="${WASM}" \
		--build-arg ARCH="${ARCH}" \
		-t axelarnet/axelar-core-${SUFFIX}:${SEMVER} --provenance=false .


.PHONY: build-push-docker-image-rosetta
build-push-docker-images-rosetta: populate-bytecode guard-SEMVER
	@DOCKER_BUILDKIT=1 docker buildx build -f Dockerfile.rosetta \
		--platform linux/amd64 \
		--output "type=image,push=${PUSH_DOCKER_IMAGE}" \
		--build-arg WASM="${WASM}" \
		-t axelarnet/axelar-core:${SEMVER}-rosetta .


# Build a docker image that is able to run dlv and a debugger can be hooked up to
.PHONY: docker-image-debug
docker-image-debug:
	@DOCKER_BUILDKIT=1 docker build --build-arg WASM="${WASM}" -t axelar/core-debug -f ./Dockerfile.debug .

# Install all generate prerequisites
.Phony: prereqs
prereqs:
	@which goimports &>/dev/null	 ||	go install golang.org/x/tools/cmd/goimports
	@which stringer &>/dev/null		 ||	go install golang.org/x/tools/cmd/stringer
	@which moq &>/dev/null			 ||	go install github.com/matryer/moq
	@which statik &>/dev/null        ||	go install github.com/rakyll/statik
	@which mdformat &>/dev/null 	 ||	pip3 install mdformat
	@which protoc &>/dev/null 		 || echo "Please install protoc for grpc (https://grpc.io/docs/languages/go/quickstart/)"
	@which golangci-lint &>/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3

# Run all the code generators in the project
.PHONY: generate
generate: prereqs
	go generate -x ./...


.PHONE: tofnd-client
tofnd-client:
	@echo -n Generating protobufs...
	@protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/tss/tofnd/v1beta1/tofnd.proto
	@echo done

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-update-deps proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@DOCKER_BUILDKIT=1 docker build -t axelar/proto-gen -f ./Dockerfile.protocgen .
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace axelar/proto-gen sh ./scripts/protocgen.sh
	@echo "Generating Protobuf Swagger endpoint"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace axelar/proto-gen sh ./scripts/protoc-swagger-gen.sh
	@statik -src=./client/docs/static -dest=./client/docs -f -m

proto-format:
	@echo "Formatting Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace tendermintdev/docker-build-proto \
	find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@echo "Linting Protobuf files"
	@$(DOCKER_BUF) lint

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

TM_URL              	= https://raw.githubusercontent.com/cometbft/cometbft/v0.34.27/proto/tendermint
GOGO_PROTO_URL      	= https://raw.githubusercontent.com/regen-network/protobuf/cosmos
GOOGLE_PROTOBUF_URL		= https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf
GOOGLE_API_URL			= https://raw.githubusercontent.com/googleapis/googleapis/master/google/api
COSMOS_PROTO_URL    	= https://raw.githubusercontent.com/regen-network/cosmos-proto/master
CONFIO_URL          	= https://raw.githubusercontent.com/confio/ics23/go/v0.9.0

TM_CRYPTO_TYPES     	= third_party/proto/tendermint/crypto
TM_ABCI_TYPES       	= third_party/proto/tendermint/abci
TM_TYPES            	= third_party/proto/tendermint/types
TM_VERSION          	= third_party/proto/tendermint/version
TM_LIBS             	= third_party/proto/tendermint/libs/bits
TM_P2P              	= third_party/proto/tendermint/p2p

GOGO_PROTO_TYPES    	= third_party/proto/gogoproto
GOOGLE_API_TYPES		= third_party/proto/google/api
GOOGLE_PROTOBUF_TYPES	= third_party/proto/google/protobuf
COSMOS_PROTO_TYPES  	= third_party/proto/cosmos_proto
# For some reason ibc expects confio proto files to be in the main folder
CONFIO_TYPES        	= third_party/proto

proto-update-deps:
	@echo "Updating Protobuf deps"
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(GOOGLE_API_TYPES)
	@curl -sSL $(GOOGLE_API_URL)/annotations.proto > $(GOOGLE_API_TYPES)/annotations.proto
	@curl -sSL $(GOOGLE_API_URL)/http.proto > $(GOOGLE_API_TYPES)/http.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

## Importing of tendermint protobuf definitions currently requires the
## use of `sed` in order to build properly with cosmos-sdk's proto file layout
## (which is the standard Buf.build FILE_LAYOUT)
## Issue link: https://github.com/tendermint/tendermint/issues/5021
	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_VERSION)
	@curl -sSL $(TM_URL)/version/types.proto > $(TM_VERSION)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto
	@curl -sSL $(TM_URL)/types/evidence.proto > $(TM_TYPES)/evidence.proto
	@curl -sSL $(TM_URL)/types/params.proto > $(TM_TYPES)/params.proto
	@curl -sSL $(TM_URL)/types/validator.proto > $(TM_TYPES)/validator.proto
	@curl -sSL $(TM_URL)/types/block.proto > $(TM_TYPES)/block.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

	@mkdir -p $(TM_LIBS)
	@curl -sSL $(TM_URL)/libs/bits/types.proto > $(TM_LIBS)/types.proto

	@mkdir -p $(TM_P2P)
	@curl -sSL $(TM_URL)/p2p/types.proto > $(TM_P2P)/types.proto

	@mkdir -p $(CONFIO_TYPES)
	@curl -sSL $(CONFIO_URL)/proofs.proto > $(CONFIO_TYPES)/proofs.proto
## insert go package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32
	@./scripts/sed.sh $(CONFIO_TYPES)/proofs.proto

	@./scripts/proto-copy-cosmos-sdk.sh

.PHONY: proto-all proto-gen proto-gen-any proto-format proto-lint proto-check-breaking proto-update-deps

guard-%:
	@ if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi
