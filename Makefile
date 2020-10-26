PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Axelar \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=axelard \
	-X github.com/cosmos/cosmos-sdk/version.ClientName=axelarcli \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) 

BUILD_FLAGS := -ldflags '$(ldflags)'

.PHONY: all
all: install docker-image

.PHONY: install
install: go.sum
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/axelarD
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/axelarCLI

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

.PHONY: build
build: go.sum
		go build -o ./bin/axelard -mod=readonly $(BUILD_FLAGS) ./cmd/axelarD
		go build -o ./bin/axelarcli -mod=readonly $(BUILD_FLAGS) ./cmd/axelarCLI

.PHONY: debug
debug: go.sum
		go build -o ./bin/axelard -mod=readonly $(BUILD_FLAGS) -gcflags="all=-N -l" ./cmd/axelarD
		go build -o ./bin/axelarcli -mod=readonly $(BUILD_FLAGS) -gcflags="all=-N -l" ./cmd/axelarCLI

.PHONY: docker-image
docker-image:
	@docker build -t axelar/core .

.PHONY: docker-image-debug
docker-image-debug:
	@docker build -t axelar/core-debug -f ./Dockerfile.debug .

.PHONY: copy-tssd
copy-tssd:
	@rsync -ru ../tssd ./

