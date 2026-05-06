# Changelog

## [Unreleased]

## [v1.4.5](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.4.5)

### State Machine Breaking

* [#2321](https://github.com/axelarnetwork/axelar-core/pull/2321) Remove link-deposit protocol and streamline cross-chain messaging — removes TSS module, deposit confirmation flow, and transfer rate limiting

### Features

* [#2334](https://github.com/axelarnetwork/axelar-core/pull/2334) Add wasm fallback for unregistered destination chains in EVM `EndBlocker`

### Bug Fixes

* [#2330](https://github.com/axelarnetwork/axelar-core/pull/2330) Allow correct decoding of historical `RotateKeyRequest` transactions
* [#2326](https://github.com/axelarnetwork/axelar-core/pull/2326) Reject votes on completed polls after grace period
* [#2335](https://github.com/axelarnetwork/axelar-core/pull/2335) Re-add chain activation check on outgoing IBC transfers
* [#2325](https://github.com/axelarnetwork/axelar-core/pull/2325) Handle historical transactions with deprecated fields and UTF-8 issues in Rosetta
* [#2322](https://github.com/axelarnetwork/axelar-core/pull/2322) Fix temp directory not being cleaned up after use
* Reject votes on failed polls
* Fix nexus routing and `RetryFailedEvent` bugs
* Cache tallied votes in `x/vote` to fix O(N²) storage reads in `EndBlocker` (Immunefi bug bounty report #62661)
* Re-lock tokens to escrow on `EndBlocker` IBC transfer failure so `RetryIBCTransfer` can recover stranded funds (Immunefi bug bounty reports #63113, #63746)
* Handle `MsgSendOperation` with fee operations in Rosetta
* Reduce `maxBitmapSize` from 32,768 to 1,024 in nexus `MaintainerState` to lower memory footprint of vote-tracking bitmaps (`MissingVotes`, `IncorrectVotes`)
* Resolve `wasmvm` version dynamically from `go.sum` in Docker builds (fixes `undefined reference to sync_pinned_codes` linker error after wasmd v0.54.6 bump)
* Fix v1.4 upgrade handler to omit re-added `crisis` and `consensus` stores that caused IAVL to reject pre-existing data and crash nodes during upgrade

### Improvements

* Update CometBFT to v0.38.22 (includes [CSA-2026-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-c32p-wcqj-j677) fix)
* Bump wasmd v0.54.3 → v0.54.7 (security fix [CWA-2026-001](https://github.com/CosmWasm/advisories)), wasmvm v2.2.4 → v2.2.6, cosmos-sdk v0.50.14 → v0.50.15, ibc-go v8.6.1 → v8.8.0
* [#2319](https://github.com/axelarnetwork/axelar-core/pull/2319) Remove time-based activation gate for validator reward fix
* [#2333](https://github.com/axelarnetwork/axelar-core/pull/2333) Bump `bytedance/sonic` to v1.15.0 for Go 1.26 compatibility
* [#2299](https://github.com/axelarnetwork/axelar-core/pull/2299) Use standard `crypto/sha3` instead of `golang.org/x/crypto/sha3`

## [v1.3.11](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.11)

### Improvements

* Update CometBFT dependency to v0.38.22

## [v1.3.10](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.10)

### Bug Fixes

* [#2329](https://github.com/axelarnetwork/axelar-core/pull/2329) Allow correct decoding of historical `RotateKeyRequest` transactions

## [v1.3.9](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.9)

### Improvements

* [#2328](https://github.com/axelarnetwork/axelar-core/pull/2328) Update CometBFT dependency to public release (includes fix for [CSA-2026-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-c32p-wcqj-j677))

## [v1.3.8](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.8)

### Bug Fixes

* [#2325](https://github.com/axelarnetwork/axelar-core/pull/2325) Handle historical transactions with deprecated fields and UTF-8 issues in Rosetta
* [#2322](https://github.com/axelarnetwork/axelar-core/pull/2322) Fix temp directory not being cleaned up after use

### Improvements

* Update CometBFT with fix for [CSA-2026-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-c32p-wcqj-j677)
* [#2299](https://github.com/axelarnetwork/axelar-core/pull/2299) Use standard `crypto/sha3` instead of `golang.org/x/crypto/sha3`

## [v1.3.6](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.6)

### Improvements

* [#2317](https://github.com/axelarnetwork/axelar-core/pull/2317) Update rosetta dependency with sub-account balance queries and memo support in transaction metadata

## [v1.3.5](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.3.5)

### Bug Fixes

* Fix external chain voting inflation rewards not being distributed to chain maintainers

### Improvements

* Update rosetta dependency

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

## [v1.2.4](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.2.4)

### Improvements

* [#2277](https://github.com/axelarnetwork/axelar-core/pull/2277) Upgrade go-ethereum from v1.10.26 to v1.16.5
* [#2278](https://github.com/axelarnetwork/axelar-core/pull/2278) Update Go version from 1.23 to 1.24 in Dockerfiles

## [v1.2.3](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.2.3)

### Bug Fixes

* Update wasmd version to fix calldepth issue

## [v1.2.2](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.2.2)

### Bug Fixes

* Fix proposal execution, wasmd call-depth, and cometbft issues via dependency updates

## [v1.2.1](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.2.1)

### State Machine Breaking

* [#2242](https://github.com/axelarnetwork/axelar-core/pull/2242) Add burner permission to distribution module account

## [v1.2.0](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.2.0)

### State Machine Breaking

* [#2231](https://github.com/axelarnetwork/axelar-core/pull/2231) Custom fee allocation - keeps community pool portion and burns the rest
* [#2236](https://github.com/axelarnetwork/axelar-core/pull/2236) Bump CosmWasm dependencies

### Bug Fixes

* [#2234](https://github.com/axelarnetwork/axelar-core/pull/2234) Use wrapped keeper for distribution begin blocker
* [#2211](https://github.com/axelarnetwork/axelar-core/pull/2211) Fix valid decimal range
* [#2209](https://github.com/axelarnetwork/axelar-core/pull/2209) Fix migration to use module name instead of module account address

### Improvements

* [#2232](https://github.com/axelarnetwork/axelar-core/pull/2232) Mark multisig sender field as deprecated and add new field with proper type

## [v1.1.3](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.1.3)

### Improvements

* [#2218](https://github.com/axelarnetwork/axelar-core/pull/2218) Update SDK dependencies

## [v1.1.2](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.1.2)

### Bug Fixes

* [#2212](https://github.com/axelarnetwork/axelar-core/pull/2212) Fix decimal range validation (ASA-2024-010 security fix)

## [v1.1.1](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.1.1)

### Bug Fixes

* [#2210](https://github.com/axelarnetwork/axelar-core/pull/2210) Fix axelarnet migration to use module name instead of module account address in SendCoinsFromModuleToModule

## [v1.1.0](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.1.0)

### State Machine Breaking

* [#2186](https://github.com/axelarnetwork/axelar-core/pull/2186), [#2202](https://github.com/axelarnetwork/axelar-core/pull/2202) Refactor IBC transfer handling - move coin locking to nexus module and improve refund flow
* [#2179](https://github.com/axelarnetwork/axelar-core/pull/2179) Require Go 1.23
* [#2182](https://github.com/axelarnetwork/axelar-core/pull/2182), [#2178](https://github.com/axelarnetwork/axelar-core/pull/2178) Expose chain registration and transaction info queries to wasm contracts

### Features

* [#2175](https://github.com/axelarnetwork/axelar-core/pull/2175) Support CallContractWithToken from nexus gateway contract
* [#2199](https://github.com/axelarnetwork/axelar-core/pull/2199) Add metadata to GMP events
* [#2173](https://github.com/axelarnetwork/axelar-core/pull/2173) Reduce heartbeat gas costs by removing key id check

### Bug Fixes

* [#2208](https://github.com/axelarnetwork/axelar-core/pull/2208), [#2203](https://github.com/axelarnetwork/axelar-core/pull/2203) Fix IBC transfer retry functionality
* [#2194](https://github.com/axelarnetwork/axelar-core/pull/2194) Fix coin type detection for external cosmos chain transfers
* [#2192](https://github.com/axelarnetwork/axelar-core/pull/2192) Fix wasm type interface conversion
* [#2169](https://github.com/axelarnetwork/axelar-core/pull/2169) Ignore malformed EVM events without topics

## [v1.0.5](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.5)

### Bug Fixes

* [#2213](https://github.com/axelarnetwork/axelar-core/pull/2213) Fix decimal range validation (ASA-2024-010 security fix)

## [v1.0.4](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.4)

### Bug Fixes

* [#2191](https://github.com/axelarnetwork/axelar-core/pull/2191) Fix vald EVM type conversion ambiguity

## [v1.0.3](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.3)

### Features

* [#2189](https://github.com/axelarnetwork/axelar-core/pull/2189) Add metadata to GMP events

## [v1.0.2](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.2)

### Features

* [#2174](https://github.com/axelarnetwork/axelar-core/pull/2174) Reduce heartbeat gas costs by removing key id check

## [v1.0.1](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.1)

### Bug Fixes

* [#2170](https://github.com/axelarnetwork/axelar-core/pull/2170) Ignore malformed EVM events without topics

## [v1.0.0](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.0.0)

### State Machine Breaking

* [#2168](https://github.com/axelarnetwork/axelar-core/pull/2168) Allow routing messages from gov module to wasm
* [#2152](https://github.com/axelarnetwork/axelar-core/pull/2152) Pass message ID between amplifier and core
* [#2145](https://github.com/axelarnetwork/axelar-core/pull/2145) Update to latest IBC-go patch
* [#2139](https://github.com/axelarnetwork/axelar-core/pull/2139) Allow refundable messages to become batched

### Features

* [#2166](https://github.com/axelarnetwork/axelar-core/pull/2166) Add access control command to activate/deactivate wasm connection
* [#2140](https://github.com/axelarnetwork/axelar-core/pull/2140) Use BatchRequest in vald to allow ignoring failed message execution

### Bug Fixes

* [#2163](https://github.com/axelarnetwork/axelar-core/pull/2163) Enable CosmWasm 1.1 and 1.2 capabilities
* [#2161](https://github.com/axelarnetwork/axelar-core/pull/2161) Allow CosmWasm client to store larger contract bytecodes
* [#2156](https://github.com/axelarnetwork/axelar-core/pull/2156) Allow incoming messages from IBC to be forwarded to wasm
* [#2155](https://github.com/axelarnetwork/axelar-core/pull/2155) Replace native asset with bond denom for dust amount
