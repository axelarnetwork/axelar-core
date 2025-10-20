PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := "1.3.0"
COMMIT := $(shell git log -1 --format='%H')

DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
HTTPS_GIT := https://github.com/axelarnetwork/axelar-core.git
PUSH_DOCKER_IMAGE := true

# Default values that can be overridden by the caller via `make VAR=value [target]`
# NOTE: Avoid adding comments on the same line as the variable assignment since trailing spaces will be included in the variable by make
WASM := true
# 3 MB max wasm bytecode size
MAX_WASM_SIZE := $(shell echo "$$((3 * 1024 * 1024))")
IBC_WASM_HOOKS := false
# Export env var to go build so Cosmos SDK can see it
export CGO_ENABLED := 1

$(info $$WASM is [${WASM}])
$(info $$IBC_WASM_HOOKS is [${IBC_WASM_HOOKS}])
$(info $$MAX_WASM_SIZE is [${MAX_WASM_SIZE}])
$(info $$CGO_ENABLED is [${CGO_ENABLED}])

ifndef $(WASM_CAPABILITIES)
# Wasm capabilities: https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
WASM_CAPABILITIES := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3"
else
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
	-X github.com/axelarnetwork/axelar-core/x/axelarnet/exported.NativeAsset=$(DENOM) \
	-X github.com/axelarnetwork/axelar-core/app.WasmEnabled=$(WASM) \
	-X github.com/axelarnetwork/axelar-core/app.IBCWasmHooksEnabled=$(IBC_WASM_HOOKS) \
	-X github.com/axelarnetwork/axelar-core/app.WasmCapabilities=$(WASM_CAPABILITIES) \
	-X github.com/axelarnetwork/axelar-core/app.MaxWasmSize=${MAX_WASM_SIZE} \
	-w -s ${STATIC_LINK_FLAGS}"

BUILD_FLAGS := -tags $(BUILD_TAGS) -ldflags $(ldflags) -trimpath
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
OS := $(shell echo $$OS_TYPE | sed -e 's/ubuntu-22.04/linux/; s/macos-latest/darwin/')
SUFFIX := $(shell echo $$PLATFORM | sed 's/\//-/' | sed 's/\///')

.PHONY: all
all: generate goimports lint build docker-image docker-image-debug

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

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
		--build-arg IBC_WASM_HOOKS="${IBC_WASM_HOOKS}" \
		--build-arg ARCH="${ARCH}" \
		-t axelar/core .

# Build a release image
.PHONY: docker-image-local-user
docker-image-local-user:  guard-VERSION guard-GROUP_ID guard-USER_ID
	@DOCKER_BUILDKIT=1 docker build \
		--build-arg USER_ID=${USER_ID} \
		--build-arg GROUP_ID=${GROUP_ID} \
		--build-arg WASM="${WASM}" \
		--build-arg IBC_WASM_HOOKS="${IBC_WASM_HOOKS}" \
		--build-arg ARCH="${ARCH}" \
		-t axelarnet/axelar-core:${VERSION}-local .

.PHONY: build-push-docker-image
build-push-docker-images:  guard-SEMVER
	@DOCKER_BUILDKIT=1 docker buildx build \
		--platform ${PLATFORM} \
		--output "type=image,push=${PUSH_DOCKER_IMAGE}" \
		--build-arg WASM="${WASM}" \
		--build-arg IBC_WASM_HOOKS="${IBC_WASM_HOOKS}" \
		--build-arg ARCH="${ARCH}" \
		-t axelarnet/axelar-core-${SUFFIX}:${SEMVER} --provenance=false .


.PHONY: build-push-docker-image-rosetta
build-push-docker-images-rosetta: populate-bytecode guard-SEMVER
	@DOCKER_BUILDKIT=1 docker buildx build -f Dockerfile.rosetta \
		--platform linux/amd64 \
		--output "type=image,push=${PUSH_DOCKER_IMAGE}" \
		--build-arg WASM="${WASM}" \
		--build-arg IBC_WASM_HOOKS="${IBC_WASM_HOOKS}" \
		-t axelarnet/axelar-core:${SEMVER}-rosetta .


# Build a docker image that is able to run dlv and a debugger can be hooked up to
.PHONY: docker-image-debug
docker-image-debug:
	@DOCKER_BUILDKIT=1 docker build --build-arg WASM="${WASM}" --build-arg IBC_WASM_HOOKS="${IBC_WASM_HOOKS}" -t axelar/core-debug -f ./Dockerfile.debug .

# Install all generate prerequisites
.Phony: prereqs
prereqs:
	@which mdformat &>/dev/null 	 ||	pip3 install mdformat
	@which protoc &>/dev/null 		 || echo "Please install protoc for grpc (https://grpc.io/docs/languages/go/quickstart/)"
	go install golang.org/x/tools/cmd/goimports@v0.38.0
	go install golang.org/x/tools/cmd/stringer@v0.38.0
	go install github.com/matryer/moq
	go install github.com/rakyll/statik
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.0

# Run all the code generators in the project
.PHONY: generate
generate: prereqs docs generate-mocks

.PHONY: generate-mocks
generate-mocks:
	go generate -x ./...

.PHONY: docs
docs:
	@echo "Removing old clidocs"

	@if find docs/cli -name "*.md"  | grep -q .; then \
		rm docs/cli/*.md; \
	fi

	@echo "Generating new cli docs"
	@go run $(BUILD_FLAGS) cmd/axelard/main.go --docs docs/cli
	@# ensure docs are canonically formatted
	@mdformat docs/cli/*

.PHONE: tofnd-client
tofnd-client:
	@echo -n Generating protobufs...
	@protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/tss/tofnd/v1beta1/tofnd.proto
	@echo done

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-format proto-lint proto-gen proto-swagger-gen

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run -u 0 --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@make clean
	@echo "Downloading Protobuf dependencies"
	@make proto-swagger-download-deps
	@echo "Generating Protobuf Swagger endpoint"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@echo "Linting Protobuf files"
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

###############################################################################
#                             Swagger Configuration                           #
###############################################################################

SWAGGER_DIR=./swagger-proto
THIRD_PARTY_DIR=$(SWAGGER_DIR)/third_party

# Dependency versions
COSMOS_SDK_VERSION=$(shell go list -m -json github.com/cosmos/cosmos-sdk | jq -r .Version)
IBC_VERSION=$(shell go list -m all | grep "github.com/cosmos/ibc-go/v" | cut -d' ' -f2)
WASMD_VERSION=$(shell go list -m -json github.com/CosmWasm/wasmd |  jq -r .Version)

proto-swagger-download-deps:
	@make clean
	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-sdk.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
    git fetch origin ${COSMOS_SDK_VERSION} && \
    git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	cd "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/ibc-go.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git fetch origin ${IBC_VERSION} && \
    git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/ibc_tmp"


	mkdir -p "$(THIRD_PARTY_DIR)/wasmd_tmp" && \
	cd "$(THIRD_PARTY_DIR)/wasmd_tmp" && \
	git init && \
	git remote add origin "https://github.com/CosmWasm/wasmd.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git fetch origin ${WASMD_VERSION} && \
    git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/wasmd_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-proto.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull origin main && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_proto_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos/ics23/v1" && \
	curl -sSL https://raw.githubusercontent.com/cosmos/ics23/master/proto/cosmos/ics23/v1/proofs.proto > "$(THIRD_PARTY_DIR)/cosmos/ics23/v1/proofs.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/gogoproto" && \
	curl -SSL https://raw.githubusercontent.com/cosmos/gogoproto/main/gogoproto/gogo.proto > "$(THIRD_PARTY_DIR)/gogoproto/gogo.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/google/api" && \
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > "$(THIRD_PARTY_DIR)/google/api/annotations.proto"
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > "$(THIRD_PARTY_DIR)/google/api/http.proto"

clean:
	rm -rf \
    tmp-swagger-gen/ \
	swagger-proto


.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking proto-swagger-download-deps clean

guard-%:
	@ if [ -z '${${*}}' ]; then echo 'Environment variable $* not set' && exit 1; fi
