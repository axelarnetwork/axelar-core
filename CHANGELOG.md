# Changelog

## [Unreleased]

### State Machine Breaking

* [#2253](https://github.com/axelarnetwork/axelar-core/pull/2253) Upgrade to Cosmos SDK v0.50 - see [Cosmos SDK v0.50 CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/CHANGELOG.md)
* [#2253](https://github.com/axelarnetwork/axelar-core/pull/2253) Upgrade to IBC v8 - see [IBC v8 CHANGELOG](https://github.com/cosmos/ibc-go/blob/release/v8.6.x/CHANGELOG.md)
* [#2253](https://github.com/axelarnetwork/axelar-core/pull/2253) Upgrade to CosmWasm v0.54.3 with wasmvm v2.2.4 - see [wasmd CHANGELOG](https://github.com/CosmWasm/wasmd/blob/v0.54.3/CHANGELOG.md)
* [#2279](https://github.com/axelarnetwork/axelar-core/pull/2279) (axelarnet, evm, multisig, nexus, permission, reward, snapshot, tss, vote) Add `MsgUpdateParams` for governance-controlled parameter updates
* (axelarnet) Increase `RouteTimeoutWindow` from 17,000 to 85,000 blocks for 1s block time
* (evm) Update `VotingGracePeriod` to 15 blocks and `RevoteLockingPeriod` to 75 blocks for 1s block time
* (evm) Migrate contract bytecode to latest version for all EVM chains
* (multisig) Update timeout parameters for 1s block time: `KeygenTimeout`, `KeygenGracePeriod`, `SigningTimeout`, `SigningGracePeriod` to 50 blocks
* (nexus) Add new parameters: `Gateway` and `EndBlockerLimit`

### Features

* [#2279](https://github.com/axelarnetwork/axelar-core/pull/2279) Add `MsgUpdateParams` messages for all axelar modules
* [#2283](https://github.com/axelarnetwork/axelar-core/pull/2283) Enable optimistic block execution

### Improvements

* [#2281](https://github.com/axelarnetwork/axelar-core/pull/2281) Reinstate reserved proto fields as deprecated for backward compatibility
* [#2268](https://github.com/axelarnetwork/axelar-core/pull/2268) Add amino names to all messages
* [#2286](https://github.com/axelarnetwork/axelar-core/pull/2286) Add missing wasmd ante handlers
* [#2280](https://github.com/axelarnetwork/axelar-core/pull/2280) Improve wasm directory path handling
* [#2260](https://github.com/axelarnetwork/axelar-core/pull/2260) Pin GitHub Actions versions for reproducible builds
* (app) Add legacy type URL support for backward compatibility
* (app) Add regression tests for historical transaction decoding
* (ci) Pin tool versions for reproducible builds
* (ci) Use official Cosmos proto-builder image

### Bug Fixes

* [#2280](https://github.com/axelarnetwork/axelar-core/pull/2280) Fix wasm directory handling
* [#2266](https://github.com/axelarnetwork/axelar-core/pull/2266) Fix message type checking
* (axelarnet) Add missing route timeout window parameter migration

### CLI Breaking Changes

* Rename `tendermint` commands to `comet`
* Move genesis commands under `genesis` subcommand
* Change default broadcast mode from `block` to `sync`

### Client Breaking Changes

* Rename CometBFT endpoints from `/cosmos/base/tendermint/v1beta1/*` to `/cosmos/base/comet/v1beta1/*`
