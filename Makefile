PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Axelar \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=axelarD \
	-X github.com/cosmos/cosmos-sdk/version.ClientName=axelarCLI \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) 

BUILD_FLAGS := -ldflags '$(ldflags)'

.PHONY: all
all: install prometheus

.PHONY: install
install: go.sum
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/axelarD
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/axelarCLI
		go install -mod=readonly $(BUILD_FLAGS) ./cmd/testCLI

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
		go build -o ./build/axelarD -mod=readonly $(BUILD_FLAGS) ./cmd/axelarD
		go build -o ./build/axelarCLI -mod=readonly $(BUILD_FLAGS) ./cmd/axelarCLI
		go build -o ./build/testCLI -mod=readonly $(BUILD_FLAGS) ./cmd/axelarCLI

.PHONY: docker
docker:
	docker-compose -f docker-compose.build.yml up --remove-orphans

.PHONY: prometheus
prometheus:
	@if [ ! -f .env ]; then \
    	  cp .env.default .env; \
	fi
	@./docker/prometheus/prometheusSetup.sh
