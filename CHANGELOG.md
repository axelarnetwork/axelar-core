# Changelog

## [v1.5.2](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.5.2)

### State Machine Breaking

* [#2384](https://github.com/axelarnetwork/axelar-core/pull/2384) Seed the `x/feepolicy` allowed fee denoms from the staking bond denom in the v1.5 upgrade handler, instead of leaving the module's default genesis allowlist of `uaxl` on chains that bond a different denom

## [v1.5.1](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.5.1)

### State Machine Breaking

* [#2370](https://github.com/axelarnetwork/axelar-core/pull/2370) Limit a transaction to paying fees in a single denomination, taken from a governance-controlled allowlist held by the new `x/feepolicy` module (default `["uaxl"]`)
* [#2373](https://github.com/axelarnetwork/axelar-core/pull/2373) Order the nexus EVM processing-message queue by insertion sequence (FIFO) so messages are delivered in arrival order rather than by message ID
* [#2378](https://github.com/axelarnetwork/axelar-core/pull/2378) Stop queueing messages whose destination route needs the original payload, which the nexus `EndBlocker` can never supply
* [#2376](https://github.com/axelarnetwork/axelar-core/pull/2376) Resolve the completed-poll chain from the poll metadata instead of the vote result, so a result naming an unregistered chain can no longer stall the `x/vote` `EndBlocker`
* [#2375](https://github.com/axelarnetwork/axelar-core/pull/2375) Skip the missing-vote penalty when an EVM poll expires with zero votes, so it no longer marks every maintainer missing and clears their rewards
* [#2371](https://github.com/axelarnetwork/axelar-core/pull/2371) Guard the cumulative burned-fee tracker against overflow, rolling back the tracker update for that denomination instead of failing the block
* [#2381](https://github.com/axelarnetwork/axelar-core/pull/2381) Reuse an already registered EVM chain param subspace when creating a chain, instead of panicking on the attempt to register it again

## [v1.5.0](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.5.0)

Pre-release for the v1.5 upgrade. Not intended for network deployment; use v1.5.1 or later.

### State Machine Breaking

* [#2349](https://github.com/axelarnetwork/axelar-core/pull/2349) Migrate to cosmos-sdk v0.53, ibc-go v10, and wasmd v0.60. Removes the `x/crisis` and `x/capability` modules and their stores, registers the 07-tendermint light client as a modular route, and runs the ibc core (6 -> 8) and ibc transfer (5 -> 6, `DenomTrace` -> `Denom`) state migrations at the upgrade height. The ibc transfer REST/CLI query surface is renamed upstream (`denom_traces` -> `denoms`); `x/tss` remains for historical transaction decoding
* [#2358](https://github.com/axelarnetwork/axelar-core/pull/2358) Add the `x/authz` module, enabling scoped, revocable, expiring authorization grants (e.g. a validator delegating governance voting to an operational key). `MsgExec` is restricted to flat messages: it cannot wrap another `MsgExec` or a batch request, and a batch request cannot wrap a `MsgExec`
* [#2353](https://github.com/axelarnetwork/axelar-core/pull/2353) Add nexus migration (8 -> 9) that eagerly reallocates existing `MaintainerState` bitmaps down to the reduced max size, so stored states stop carrying oversized buffers in the KV store even for maintainers on deactivated chains that never vote again
* [#2359](https://github.com/axelarnetwork/axelar-core/pull/2359) Raise the wasm static validation limits (`MaxFunctionLocals` 2048, `MaxTotalFunctionLocals` 20,000) above the wasmvm v2.3.4 defaults, which reject optimizer-built amplifier contracts at store-code time
* [#2362](https://github.com/axelarnetwork/axelar-core/pull/2362) Tighten the ABI inflation guard for `CosmWasmV1` payloads: lower the inflation budget to 50kb, restrict wasm argument types to an allowlist, and charge for element types of empty dynamic arrays

### Features

* [#2344](https://github.com/axelarnetwork/axelar-core/pull/2344) Confirm transaction inclusion via CometBFT event subscriptions instead of indexer polling, so `vald` can broadcast against nodes with the transaction indexer disabled and no longer polls for every tx
* [#2361](https://github.com/axelarnetwork/axelar-core/pull/2361) Reject `MsgRetryFailedEvent` for deprecated pre-v1.4 event types with a clear error instead of a recovered panic

### Bug Fixes

* [#2356](https://github.com/axelarnetwork/axelar-core/pull/2356) Charge execution gas before translating the message payload so the decode and ABI inflation guards are metered before they run
* [#2365](https://github.com/axelarnetwork/axelar-core/pull/2365) Flatten a `RefundMsgRequest`'s inner message in the ante handler so it passes through the message ante decorators (`RestrictedTx`, `CheckProxy`, etc.), consistent with how authz `MsgExec` and auxiliary `BatchRequest` inner messages are handled
* [#2341](https://github.com/axelarnetwork/axelar-core/pull/2341) Reject `RefundMsg` requests whose inner message has a different `permission_role` than the wrapping request
* [#2346](https://github.com/axelarnetwork/axelar-core/pull/2346) Derive the refund sender from each msg, not the first
* [#2351](https://github.com/axelarnetwork/axelar-core/pull/2351) Clear accrued rewards when a chain maintainer voluntarily deregisters, matching the automatic removal path

### Improvements

* [#2349](https://github.com/axelarnetwork/axelar-core/pull/2349) Bump wasmd to v0.60.8 and wasmvm to v2.3.4, patching CWA-2026-005 (high-severity block-production delay via wasm execution)
* [#2377](https://github.com/axelarnetwork/axelar-core/pull/2377) Bump the cosmos-sdk fork to v0.53.8 and cometbft to v0.38.25
* [#2363](https://github.com/axelarnetwork/axelar-core/pull/2363) Bump the `github.com/cosmos/rosetta` fork to the `axelar-core-v1.5.x-compatible` branch (commit c9ce423) for compatibility with the v1.5 cosmos-sdk v0.53 / ibc-go v10 stack
* [#2355](https://github.com/axelarnetwork/axelar-core/pull/2355) Bump grpc (v1.82.0), x/net (v0.56.0), go-ethereum (v1.16.9), xz (v0.5.15), and go-getter (v1.7.9) to address security advisories
* [#2348](https://github.com/axelarnetwork/axelar-core/pull/2348) Bump ledger-cosmos-go to v0.15.0 and zondax/ledger-go to v1.0.0 (plus transitive dependency updates) so the CLI can talk to devices running the latest Ledger firmware and Cosmos app
* [#2369](https://github.com/axelarnetwork/axelar-core/pull/2369) Disable `axelard export`: state export to genesis is not supported, since axelar-core upgrades via in-place store migrations rather than genesis export/import. The command now returns a clear error instead of silently producing non-round-trippable genesis (in-flight IBC correlation, the nexus processing-message index, the wasm activation flag, and in-flight vote tallies are not round-trippable)
* [#2367](https://github.com/axelarnetwork/axelar-core/pull/2367) Remove the dead link-deposit CLI commands (`evm link`, `evm confirm-erc20-deposit`, `evm create-burn-tokens`, `axelarnet link`, `axelarnet confirm-deposit`) whose backing Msg RPCs were removed in [#2321](https://github.com/axelarnetwork/axelar-core/pull/2321)
* [#2360](https://github.com/axelarnetwork/axelar-core/pull/2360) Remove the superseded nexus (6 -> 7, 7 -> 8) and axelarnet (7 -> 8) in-place store migrations and their registrations. Consensus versions are unchanged (nexus 9, axelarnet 8)
* [#2364](https://github.com/axelarnetwork/axelar-core/pull/2364) Remove the superseded multisig (2 -> 3), permission (1 -> 2), reward (1 -> 2), tss (3 -> 4), vote (2 -> 3), and evm (10 -> 11 link-deposit cleanup) in-place store migrations and their registrations. All networks store these modules at their current consensus versions, so the handlers can no longer be invoked. The evm bytecode migration stays registered wrapping a no-op
* [#2343](https://github.com/axelarnetwork/axelar-core/pull/2343) Speed up block catch-up by persisting the reward pool once per batch instead of re-marshaling and writing it for every reward added
* [#2342](https://github.com/axelarnetwork/axelar-core/pull/2342) Speed up block catch-up by recovering maintainer addresses directly from store keys instead of unmarshaling each stored value
* [#2350](https://github.com/axelarnetwork/axelar-core/pull/2350) Add regression tests pinning the per-source `CommandID` (EVM, Cosmos, amplifier) for the nexus to EVM message delivery route

## [v1.4.9](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.4.9)

### Bug Fixes

* [#2357](https://github.com/axelarnetwork/axelar-core/pull/2357) Broadcast an empty vote in `vald` when a gateway tx confirmation yields more than `MaxEventsPerVote` events, so the poll completes and maintainer rewards are not cleared

## [v1.4.8](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.4.8)

### Bug Fixes

* [#2352](https://github.com/axelarnetwork/axelar-core/pull/2352) Guard against super-linear ABI amplification when routing a MsgRouteMessage with a malformed payload

## [v1.4.7](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.4.7)

### Bug Fixes

* Restore `InterfaceRegistry` registrations for `axelar.tss.v1beta1.UpdateParamsRequest` (`sdk.Msg`) and `axelar.tss.v1beta1.HeartBeatRequest` (`sdk.Msg` + `reward.Refundable`), fixing `/cosmos/gov/v1/proposals` returning `no concrete type registered for type URL` on chains that have a historical TSS `MsgUpdateParams` proposal in gov state

## [v1.4.6](https://github.com/axelarnetwork/axelar-core/releases/tag/v1.4.6)

### Improvements

* Update CometBFT to v0.38.23

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
