# Emergency pause and halt runbook

Two emergency controls exist, both in the `permission` module and both gated on `ROLE_ACCESS_CONTROL`
(the governance multisig key). A gov proposal can also execute them, as the gov module account.

| Control | What it does | Undo |
|---|---|---|
| **Pause** (`axelard tx permission pause`) | Blocks the chain's transaction input: only allowlisted messages are accepted. Blocks keep being produced. | `axelard tx permission unpause` |
| **Halt** (`axelard tx permission halt-chain`) | Schedules an upgrade plan at the next height. Every node stops before executing it. | Restart the nodes (see [Resuming after a halt](#resuming-after-a-halt)) |

Pause is a soft halt, meant for hours, not days (see [Things to watch](#things-to-watch-while-paused)). Escalate
to a halt if block production itself must stop.

## Sending the transactions

The `ROLE_ACCESS_CONTROL` account is the governance multisig key, so every command below is generated, signed by
the key holders and broadcast as a multisig tx:

```bash
# 1. generate the unsigned tx (from the multisig key name or address)
axelard tx permission pause "<reason>" --level all --from <governance-multisig> --generate-only > tx.json

# 2. every signer signs
axelard tx sign tx.json --multisig <governance-multisig-address> --from <signer> --output-document sig-<signer>.json

# 3. combine and broadcast
axelard tx multisign tx.json <governance-multisig> sig-*.json > signed.json
axelard tx broadcast signed.json
```

The `reason` argument is required. It is stored in the emitted event, and in the upgrade plan's `info` for a halt.

## Pausing and unpausing

```bash
axelard tx permission pause "<reason>" --level all                # strictest level (default)
axelard tx permission pause "<reason>" --level except-validators
axelard tx permission unpause "<reason>"
axelard query permission is-paused                                # {"paused": true, "level": "PAUSE_LEVEL_ALL"}
```

Running `pause` again with another level changes the level of the ongoing pause. Pausing at the current level fails.

### Levels and allowlist

The allowlist is `EmergencyPauseAllowlist` in `x/permission/exported/pause.go`.

| Message | `all` | `except-validators` |
|---|:-:|:-:|
| `permission.PauseRequest`, `UnpauseRequest`, `HaltChainRequest` | ✓ | ✓ |
| `permission.UpdateGovernanceKeyRequest`, `RegisterControllerRequest`, `DeregisterControllerRequest` | | ✓ |
| `vote.VoteRequest` | | ✓ |
| `multisig.SubmitPubKeyRequest`, `multisig.SubmitSignatureRequest` | | ✓ |
| `reward.RefundMsgRequest`, `auxiliary.BatchRequest` (vald's wrappers) | | ✓ (only if every inner message is allowed) |
| everything else | | |

Every message is checked, wherever it comes from:

- top-level tx messages, in `CheckTx` (they never enter the mempool and pay no fee) and in block execution;
- messages unwrapped from `BatchRequest`, `RefundMsgRequest` and authz `MsgExec`;
- messages dispatched by CosmWasm contracts, including nexus custom messages;
- messages executed by a gov proposal.

A rejected message fails with `chain is paused` (codespace `permission`, code 3).

### Events

| Event | Attributes |
|---|---|
| `emergency_pause` | `signer`, `reason`, `height`, `pause_level`, `previous_pause_level` |
| `emergency_unpause` | `signer`, `reason`, `height`, `previous_pause_level` |
| `emergency_halt` | `signer`, `reason`, `height`, `halt_height`, `plan_name`, plus `replaced_plan_name`, `replaced_plan_height` and `replaced_plan_info` when a pending plan was overridden |

### What keeps running while paused

The pause only stops new input. Every BeginBlocker and EndBlocker still runs, so anything already queued on chain
keeps going:

- **nexus**: routes queued general messages, including to the wasm gateway, and runs the chain maintainer checks.
- **axelarnet**: sends queued IBC transfers.
- **evm**: processes already confirmed events and creates command batches. Batches already signed can still be
  relayed to EVM chains.
- **vote** and **multisig**: open polls and keygen or signing sessions expire as usual.
- **reward, mint, distribution, staking, slashing**: inflation, rewards, unbonding and downtime jailing continue.
  Downtime jailing only depends on block signing, which continues.
- **gov**: proposals whose voting period ends during the pause are tallied. Their messages run through the pause
  check, so a passed proposal fails permanently unless all its messages are allowlisted. A proposal containing
  `UnpauseRequest` would unpause the chain. New proposals, deposits and votes are blocked.

To also freeze cross-chain flows that are already queued, deactivate the affected chains separately
(`axelard tx nexus deactivate-chain`, use `:all:` for every chain).

### Things to watch while paused

- **Validator rewards and maintainer status.** At level `all`, vald's votes and signatures are rejected. Polls and
  signing sessions that expire during the pause count those validators as missing. Their accrued rewards for those
  polls are cleared, and the nexus chain-maintainer check runs on the missed votes. At level `except-validators`,
  validators keep participating normally.
- **In-flight work at level `except-validators`.** Polls already open can still complete, so confirmed events get
  processed. Signing sessions already open can still complete, so command batches become executable on the
  destination chain. If that is part of the incident, use level `all` or deactivate the chain.
- **IBC light clients.** `MsgUpdateClient` is blocked. A pause longer than a client's trusting period lets the
  client expire, and recovering it needs a gov proposal.
- **vald, ampd and relayers** keep retrying rejected txs. Expect error logs.

## Halting the chain

```bash
axelard tx permission halt-chain "<reason>"
axelard tx permission halt-chain "<reason>" --override-pending-upgrade   # replaces an already scheduled plan
```

The tx is included at height `H` and schedules the upgrade plan `emergency-halt-<H+1>` at height `H+1`. At `H+1`,
before executing anything, every node:

- logs `UPGRADE "emergency-halt-<H+1>" NEEDED at height: <H+1>: <reason>`;
- writes `<home>/data/upgrade-info.json`;
- stops.

The last committed block is `H`. The halt works while the chain is paused, and the pause state is kept through the
halt.

- If an upgrade plan is already scheduled, the halt fails and names that plan, unless `--override-pending-upgrade`
  is set. An overridden plan is gone for good: the `emergency_halt` event records it, and it has to be scheduled
  again through gov after recovery.
- **Cosmovisor.** Cosmovisor reacts to `upgrade-info.json` and looks for
  `cosmovisor/upgrades/emergency-halt-<H+1>/bin`. With automatic binary download disabled and no such directory,
  it cannot switch binaries. Check how your Cosmovisor setup behaves before an incident. Running `axelard`
  directly for the restart avoids the question.

## Resuming after a halt

More than 2/3 of the voting power must restart for blocks to resume. Choose one option.

### Option A: same binary, skip the halt plan

Restart every node with the halt height:

```bash
axelard start --unsafe-skip-upgrades <H+1>
```

The node logs `UPGRADE "emergency-halt-<H+1>" SKIPPED`, clears the plan and produces block `H+1`. Without the
flag, the node stops again at the same point. Keep the flag only for this restart.

### Option B: patched binary with a handler for the halt plan

Use this when the fix needs a state migration at the halt height. Add the plan to `chainUpgrades` in
`app/upgrades.go`, like any other upgrade:

```go
{
	name: "emergency-halt-<H+1>",
	// optional: storeUpgrades, createHandler
},
```

Ship this binary to every validator and start it without `--unsafe-skip-upgrades`. The handler runs at `H+1` and
the chain continues. Do not start the patched binary before the chain has halted: a node that has a handler for
a plan not yet due stops with `BINARY UPDATED BEFORE TRIGGER`. Remove the entry in the next regular release.

### After resuming

If the chain was paused before the halt, it is still paused. Check with `axelard query permission is-paused`
and unpause once safe.

## Testing

`app/emergency_test.go` runs a single-validator chain in-process through ABCI with real signed transactions. It
covers:

- the pause and unpause flow;
- the allowlist at each level;
- the nested execution paths;
- halting at `H+1`;
- resuming with `--unsafe-skip-upgrades`.

There is no multi-node harness in this repo. Before relying on this on a live network, rehearse on a local
multi-node testnet:

1. pause, confirm a bank transfer fails, unpause, confirm it succeeds;
2. halt, confirm every node stops with last height `H`;
3. restart every node with `--unsafe-skip-upgrades <H+1>` and confirm blocks are produced.
