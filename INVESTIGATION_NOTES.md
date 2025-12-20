# Link-Deposit Protocol Removal - Investigation Notes

This document tracks important decisions and test plans for the link-deposit protocol removal work.

## Current Session Progress

**Test file:** `x/evm/abci_routing_test.go`

**All tests complete! (39 tests total)**

**Routing tests (TestProcessConfirmedEvents): 29 tests**
- ✅ 1.1 ContractCall (8 tests)
- ✅ 1.2 ContractCallWithToken (8 tests)
- ✅ 1.3 TokenDeployed (3 tests)
- ✅ 1.4 KeyRotation (5 tests)
- ✅ 1.5 Resilience (2 tests)
- ✅ 1.6 BoundedComputation (1 test)
- ✅ 1.7 UnsupportedEventTypes (2 tests)

**Delivery tests (TestDeliverPendingMessages): 10 tests**
- ✅ 2.1 DeliverMessage (5 tests)
- ✅ 2.2 DeliverMessageWithToken (2 tests)
- ✅ 2.3 Resilience (2 tests)
- ✅ 2.4 BoundedComputation (1 test)

## Test Implementation Notes

### Test Structure
- `routingTestSetup` struct holds all mocks: ctx, bk, n, m, sourceCk, destCk, sourceChain, destChain
- `newRoutingTestSetup(t)` creates default mocks with success behavior
- Helper methods on setup:
  - `createContractCallEvent()` - creates ContractCall event
  - `createContractCallWithTokenEvent()` - creates ContractCallWithToken event (Symbol: "AXL", Amount: 1000)
  - `createTokenDeployedEvent(symbol, tokenAddress)` - creates TokenDeployed event
  - `setupConfirmedToken()` - sets up GetERC20TokenBySymbol to return confirmed token with Asset: "uaxl"
  - `queueEvent(event)` - sets up GetConfirmedEventQueueFunc to return the event

### Key Implementation Details
- **Symbol vs Asset**: Event has `Symbol: "AXL"` (ERC20 symbol), token has `Asset: "uaxl"` (Cosmos denom)
- Tests only call `EndBlocker()` - no direct calls to internal functions
- Each test overrides only the specific mocks needed for that case
- Use `errors.New()` for test errors (requires "errors" import)
- SDK typed events use proto message name (e.g., "axelar.evm.v1beta1.ContractCallApproved")

### Imports needed in test file
```go
import (
	"errors"
	"testing"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	// ... rest of imports
)
```

### Important: Always trace implementations
When checking what a function does, **always read the actual implementation**, not just the call site. Example:
- `SetEventFailed` at call site just looks like state update
- But actual implementation in `chainKeeper.go` also emits `EVMEventFailed` event
- Similarly, `SetEventCompleted` may do more than just change state

## Test Plan for abci.go

### Structure
- Two main test functions matching EndBlocker's two operations:
  a. `TestProcessConfirmedEvents` - EVM → Nexus (routing)
  b. `TestDeliverPendingMessages` - Nexus → EVM (delivery)

### Constraints
1. Only call public functions - entry point is EndBlocker
2. Use t.Run() style, not Gherkin
3. Tests verify current behavior - should fail if behavior changes
4. Exhaustive - this is a crucial part of the system
5. Ensure EndBlocker doesn't get blocked when one event/message fails (resilience)
6. Ensure bounded computation - only processes limited amount per block
7. Test each event/message type independently - don't rely on shared code paths
8. New file for clean slate
9. **Add tests ONE AT A TIME** for easier review

### 1. Routing Tests (TestProcessConfirmedEvents) ✅ DONE (29 tests total)

#### 1.1 TestRouteContractCall ✅ DONE (8 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | event to 'Axelarnet' is marked failed | ✅ |
| 2 | event to inactive chain is marked failed | ✅ |
| 3 | event to unregistered chain is marked failed | ✅ |
| 4 | SetNewMessage error marks event failed | ✅ |
| 5 | EnqueueRouteMessage error marks event failed | ✅ |
| 6 | event to valid chain creates message and enqueues for routing | ✅ |
| 7 | routing does not create command | ✅ |
| 8 | routing does not emit ContractCallApproved | ✅ |

Note: Cases 6-10 from original plan consolidated - message content verified in success test, EnqueueRouteMessage verified in success test.

#### 1.2 TestRouteContractCallWithToken ✅ DONE (8 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | event to 'Axelarnet' is marked failed | ✅ |
| 2 | event to inactive chain is marked failed | ✅ |
| 3 | event to unregistered chain is marked failed | ✅ |
| 4 | SetNewMessage error marks event failed | ✅ |
| 5 | EnqueueRouteMessage error marks event failed | ✅ |
| 6 | source token not confirmed marks event failed | ✅ |
| 7 | event to wasm chain is marked failed | ✅ |
| 8 | event to valid chain creates message with asset and enqueues for routing | ✅ |

#### 1.3 TestApplyTokenDeployment ✅ DONE (3 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | token does not exist marks event failed | ✅ |
| 2 | token address mismatch marks event failed | ✅ |
| 3 | valid token deployment confirms token and emits event | ✅ |

#### 1.4 TestApplyKeyRotation ✅ DONE (5 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | next key ID not found marks event failed | ✅ |
| 2 | key not found marks event failed | ✅ |
| 3 | operator count mismatch marks event failed | ✅ |
| 4 | RotateKey error marks event failed | ✅ |
| 5 | valid key rotation calls RotateKey, emits event, and marks completed | ✅ |

Note: Original plan had 7 cases but tests 1-3 (success verification) consolidated into test 5.

#### 1.5 TestResilience (Routing) ✅ DONE (2 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | first event fails, second succeeds | ✅ |
| 2 | multiple chains, one fails | ✅ |

#### 1.6 TestBoundedComputation (Routing) ✅ DONE (1 test implemented)

| # | Case | Status |
|---|------|--------|
| 1 | more events than EndBlockerLimit | ✅ |

#### 1.7 TestUnsupportedEventTypes ✅ DONE (2 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | Event_TokenSent in queue | ✅ marked failed (RunCached catches panic) |
| 2 | Event_Transfer in queue | ✅ marked failed (RunCached catches panic) |

### 2. Delivery Tests (TestDeliverPendingMessages) ✅ DONE (10 tests total)

#### 2.1 TestDeliverMessage ✅ DONE (5 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | gateway not set marks message failed | ✅ |
| 2 | current key not found marks message failed | ✅ |
| 3 | destination chain deactivated marks message failed | ✅ |
| 4 | invalid contract address marks message failed | ✅ |
| 5 | valid message creates command and marks executed (includes command content + event verification) | ✅ |

#### 2.2 TestDeliverMessageWithToken ✅ DONE (2 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | destination token not confirmed marks message failed | ✅ |
| 2 | valid message with token creates command and marks executed (includes command content + event verification) | ✅ |

Note: Rate limiting tests skipped - rate limiting is being removed from EVM module (see todo #11).

#### 2.3 TestResilience (Delivery) ✅ DONE (2 tests implemented)

| # | Case | Status |
|---|------|--------|
| 1 | first message fails, second succeeds | ✅ |
| 2 | multiple chains process messages independently when one has failures | ✅ |

#### 2.4 TestBoundedComputation (Delivery) ✅ DONE (1 test implemented)

| # | Case | Status |
|---|------|--------|
| 1 | more messages than limit, rest available next block | ✅ |

## Key Behavioral Changes (old → new)

1. **ContractCall to EVM destination:**
   - Old: Directly created ApproveContractCallCommand and enqueued it, emitted ContractCallApproved
   - New: Routes to nexus via SetNewMessage + EnqueueRouteMessage, command created later during delivery

2. **ContractCallWithToken to EVM destination:**
   - Old: Validated token on destination, created command directly
   - New: Only validates token on source during routing; destination validation happens during delivery

3. **Removed event types:**
   - Event_TokenSent - completely removed
   - Event_Transfer - completely removed

4. **New: EnqueueRouteMessage** - Old code only called SetNewMessage, new code also calls EnqueueRouteMessage

## Events Emitted

### Event Placement and Rollback Safety

The pattern used ensures failure events persist while success events roll back with state if something goes wrong.

**Inside RunCached (will roll back if RunCached fails):**
- `EventTypeTokenConfirmation` - in `applyTokenDeployment`
- `EventTypeTransferKeyConfirmation` - in `applyKeyRotation`
- `ContractCallApproved` - in `deliverMessage`
- `ContractCallWithMintApproved` - in `deliverMessageWithToken`

This is **intentionally correct** - success events should only persist if state commits.

**Outside RunCached (won't roll back):**
- `EVMEventFailed` - via `SetEventFailed` in chainKeeper
- `EVMEventCompleted` - via `SetEventCompleted` in chainKeeper
- `ContractCallFailed` - directly in `deliverPendingMessages`
- `MessageFailed` - via `SetMessageFailed` in nexus keeper
- `MessageExecuted` - via `SetMessageExecuted` in nexus keeper

### Routing Phase Events (EVM event → nexus message)
| Outcome | Events |
|---------|--------|
| Token deployment success | `EventTypeTokenConfirmation` + `EVMEventCompleted` |
| Key rotation success | `EventTypeTransferKeyConfirmation` + `EVMEventCompleted` |
| ContractCall/WithToken success | `EVMEventCompleted` only |
| Any failure | `EVMEventFailed` |

### Delivery Phase Events (nexus message → EVM command)
| Outcome | Events |
|---------|--------|
| Message success | `ContractCallApproved` + `MessageExecuted` |
| MessageWithToken success | `ContractCallWithMintApproved` + `MessageExecuted` |
| Any failure | `ContractCallFailed` + `MessageFailed` |

### Retry
- `RetryFailedEvent` → emits `EVMEventRetryFailed{Chain, EventID, Type}`

## Open Questions

### Should success events be moved into keeper functions?

Currently success events (`EventTypeTokenConfirmation`, `EventTypeTransferKeyConfirmation`, `ContractCallApproved`, `ContractCallWithMintApproved`) are emitted directly in abci.go inside RunCached. This works correctly because they roll back with state on failure.

However, for consistency with failure events (which are emitted by keeper functions like `SetEventFailed`), we could consider moving success events into keeper functions like:
- `ConfirmToken()` → emit `EventTypeTokenConfirmation`
- `RotateKey()` → emit `EventTypeTransferKeyConfirmation`
- etc.

**Pros of moving:**
- Consistency: all state+event changes in one place
- Encapsulation: keeper owns both state and event emission
- Easier to test keeper functions directly

**Cons:**
- More invasive changes
- Some events are EVM-specific but keepers are in other modules (e.g., multisig.RotateKey)

**Decision:** Revisit after main cleanup is done.

## Todos

1. [in_progress] Fix abci_test.go for new function signatures
2. [pending] Remove dead code paths (Event_TokenSent, Event_Transfer) from EnqueueConfirmedEvent
3. [pending] Remove burner functions from chainKeeper.go
4. [pending] Remove deposit functions from chainKeeper.go
5. [pending] Remove ChainKeeper interface methods from expected_keepers.go
6. [pending] Deprecate Burnable param in chain params
7. [pending] Clean up genesis state fields (migration)
8. [pending] Clean up state keys (migration)
9. [pending] Add RetryFailedMessage endpoint to EVM module (also called by RetryFailedEvent for backwards compat)
10. [pending] Move axelarnet/wasm destination checks from EVM to nexus (fail during routing)
11. [pending] Remove rate limiting code from EVM abci

## Function Renames in abci.go

| Old Name | New Name | Direction/Purpose |
|----------|----------|-------------------|
| handleContractCall | routeContractCall | EVM→Nexus |
| handleContractCallWithToken | routeContractCallWithToken | EVM→Nexus |
| routeMessageToNexus | routeEventToNexus | EVM→Nexus |
| handleTokenDeployed | applyTokenDeployment | State update |
| handleMultisigTransferKey | applyKeyRotation | State update |
| handleConfirmedEvent | processConfirmedEvent | Processing |
| handleConfirmedEventsForChain | processConfirmedEventsForChain | Processing |
| handleConfirmedEvents | processConfirmedEvents | Processing |
| validateMessage | validateMessageForDelivery | Nexus→EVM |
| handleMessage | deliverMessage | Nexus→EVM |
| handleMessageWithToken | deliverMessageWithToken | Nexus→EVM |
| handleMessages | deliverPendingMessages | Nexus→EVM |
