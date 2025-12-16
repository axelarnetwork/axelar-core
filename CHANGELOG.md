# Changelog

## [Unreleased]

## [v1.3.6](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.6)

### Improvements

* [#2317](https://github.com/axelarnetwork/axelar-core/pull/2317) Update rosetta dependency with sub-account balance queries and memo support in transaction metadata

## [v1.3.5](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.5)

### Bug Fixes

* Fix validator error check in external chain voting inflation rewards - validators were incorrectly skipped due to inverted error check

### Improvements

* Update rosetta dependency to axelar-core compatible branch

## [v1.3.4](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.4)

### Bug Fixes

* [#2312](https://github.com/axelarnetwork/axelar-core/pull/2312) Fix rosetta address encoding issue

### Improvements

* [#2313](https://github.com/axelarnetwork/axelar-core/pull/2313) Deprecate vald heartbeat handler (disabled by default via `enable_heartbeat` config)

## [v1.3.3](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.3)

### Bug Fixes

* [#2310](https://github.com/axelarnetwork/axelar-core/pull/2310) Fix rosetta encoding config to include AccountI interface and apply rosetta patches

## [v1.3.2](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.2)

### Bug Fixes

* [#2304](https://github.com/axelarnetwork/axelar-core/pull/2304) Fix rosetta base64 encoding for transaction metadata

### Improvements

* [#2302](https://github.com/axelarnetwork/axelar-core/pull/2302), [#2305](https://github.com/axelarnetwork/axelar-core/pull/2305) Add statically linked linux binary and `make build-static` target

## [v1.3.1](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.1)

### Improvements

* [#2293](https://github.com/axelarnetwork/axelar-core/pull/2293) Keep IAVL fast node disabled by default (consistent with pre-upgrade behavior) to prevent unexpected re-indexing
* [#2294](https://github.com/axelarnetwork/axelar-core/pull/2294) Deprecate unused CLI commands: `axelard tx axelarnet link`, `axelard tx axelarnet confirm-deposit`, `axelard query evm token-address`
* [#2295](https://github.com/axelarnetwork/axelar-core/pull/2295), [#2296](https://github.com/axelarnetwork/axelar-core/pull/2296) Optimize vald config defaults for faster block time (shorter grace periods, faster polling)

## [v1.3.0](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.0)

### State Machine Breaking

* [#2285](https://github.com/axelarnetwork/axelar-core/pull/2285) Upgrade to Cosmos SDK v0.50, IBC v8, and CosmWasm v0.54 - see [SDK CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/CHANGELOG.md), [IBC CHANGELOG](https://github.com/cosmos/ibc-go/blob/release/v8.6.x/CHANGELOG.md), [wasmd CHANGELOG](https://github.com/CosmWasm/wasmd/blob/v0.54.3/CHANGELOG.md)
* [#2279](https://github.com/axelarnetwork/axelar-core/pull/2279) Add `MsgUpdateParams` for governance-controlled parameter updates in all Axelar modules
* [#2241](https://github.com/axelarnetwork/axelar-core/pull/2241) Add burner permission to distribution module account
* Update default module parameters for 1s block time (5x faster than before):
  * (axelarnet) `RouteTimeoutWindow`: 17,000 → 85,000 blocks
  * (evm) `VotingGracePeriod`: 3 → 15 blocks, `RevoteLockingPeriod`: 15 → 75 blocks
  * (multisig) `KeygenTimeout`, `SigningTimeout`: 10 → 50 blocks
* (evm) Migrate gateway contract bytecode to latest version for all EVM chains
* (nexus) Add `Gateway` and `EndBlockerLimit` parameters

### Features

* [#2283](https://github.com/axelarnetwork/axelar-core/pull/2283) Enable optimistic block execution for improved performance
* [#2291](https://github.com/axelarnetwork/axelar-core/pull/2291) Add governance controls to enable/disable deposit address linking per chain

### Improvements

* [#2275](https://github.com/axelarnetwork/axelar-core/pull/2275) Upgrade go-ethereum from v1.10.26 to v1.16.5
* [#2281](https://github.com/axelarnetwork/axelar-core/pull/2281) Reinstate reserved proto fields as deprecated for backward compatibility
* [#2268](https://github.com/axelarnetwork/axelar-core/pull/2268) Add amino names to all messages for Ledger signing compatibility
* [#2286](https://github.com/axelarnetwork/axelar-core/pull/2286) Add missing wasmd ante handlers
* [#2280](https://github.com/axelarnetwork/axelar-core/pull/2280) Fix wasm directory path handling

### Bug Fixes

* [#2289](https://github.com/axelarnetwork/axelar-core/pull/2289) Fix amino name for EVM LinkRequest message
* [#2290](https://github.com/axelarnetwork/axelar-core/pull/2290) Fix tm-events event filter bug that could cause missed events
* [#2266](https://github.com/axelarnetwork/axelar-core/pull/2266) Fix message type checking in ante handler

### CLI Breaking Changes

* Rename `tendermint` commands to `comet` (e.g., `axelard tendermint` → `axelard comet`)
* Move genesis commands under `genesis` subcommand
* Change default broadcast mode from `block` to `sync`

### Client Breaking Changes

* Rename CometBFT REST endpoints from `/cosmos/base/tendermint/v1beta1/*` to `/cosmos/base/comet/v1beta1/*`
