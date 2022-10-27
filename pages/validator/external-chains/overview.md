# Overview

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

As a validator for the Axelar network, your Axelar node will vote on the status of external blockchains such as Ethereum, Cosmos, etc. Specifically:

1. Select which external chains your Axelar node will support. Set up and configure your own nodes for the chains you selected.
2. Provide RPC endpoints for these nodes to your Axelar validator node and register as a maintainer for these chains on the Axelar network.

<Callout type="warning" emoji="⚠️">

For item 2 above the following tasks should always be done together:

- Enable/disable RPC endpoints.
- Register/deregister as chain maintainer.

Failure to do these tasks together could result in loss of transaction fees, loss of validator rewards, and poor validator performance.

See below for details. Read this entire article before you begin supporting external chains.

</Callout>

## External chains you can support on Axelar

- EVM-compatible chains
  - [Aurora](./aurora)
  - [Avalanche](./avalanche)
  - [Binance](./binance)
  - [Ethereum](./ethereum)
  - [Fantom](./fantom)
  - [Moonbeam](./moonbeam)
  - [Polygon](./polygon)
- Cosmos chains
  - Nothing to do. All Cosmos chains are automatically supported by default.

## Add external chain info to your validator's configuration

In the `axelarate-community` git repo edit the file `configuration/config.toml`: set the `rpc_addr` and `start-with-bridge` entries corresponding to the external chain you wish to connect.

Your `config.toml` file should already contain a snippet like the following:

```toml
##### EVM bridges options #####
# Each EVM chain needs the following
# 1. `[[axelar_bridge_evm]]` # header
# 2. `name`                  # chain name (eg. "Ethereum")
# 3. 'rpc_addr'              # EVM RPC endpoint URL; chain maintainers set their own endpoint
# 4. `start-with-bridge`     # `true` to support this chain
#
# see https://docs.axelar.dev/validator/external-chains

[[axelar_bridge_evm]]
name = "Ethereum"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "Avalanche"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "Fantom"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "Moonbeam"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "Polygon"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "binance"
rpc_addr = ""
start-with-bridge = false

[[axelar_bridge_evm]]
name = "aurora"
rpc_addr = ""
start-with-bridge = false
```

### Example: Ethereum

Edit the `Ethereum` entry::

```toml
[[axelar_bridge_evm]]
name = "Ethereum"
rpc_addr = "my_ethereum_host"
start-with-bridge = true
```

Substitute your Ethereum RPC address for `my_ethereum_host`. Be sure to set `start-with-bridge` to `true`.

## Restart your companion processes

<Callout type="warning" emoji="⚠️">
  Caution: Do not stop the `axelar-core` process. If you stop `axelar-core` then you risk downtime for Tendermint consensus, which can result in penalties.
</Callout>

<Callout type="warning" emoji="⚠️">
  Caution: If `vald`, `tofnd` are stopped for too long then your validator might fail to produce a heartbeat transaction when needed. The risk of this event can be reduced to near-zero if you promptly restart these processes shortly after a recent round of heartbeat transactions.
</Callout>

<Callout emoji="💡">
  Tip: Heartbeat events are emitted every 50 blocks. Your validator typically responds to heartbeat events within 1-2 blocks. It should be safe to restart `vald`, `tofnd` at block heights that are 5-10 mod 50.
</Callout>

Stop your companion processes `vald`, `tofnd`.

```bash
kill -9 $(pgrep tofnd)
kill -9 $(pgrep -f "axelard vald-start")
```

Immediately resume your companion processes `vald`, `tofnd` as per [Launch companion processes](../setup/vald-tofnd).

## Check your connections to new chains in vald

Check your `vald` logs to see that your validator node has successfully connected to the new EVM chains you added. [[How to view logs.]](../setup/vald-tofnd)

You should see something like:

<CodeBlock language="log">
{`2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Ethereum module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Avalanche module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Fantom module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Moonbeam module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Polygon module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Binance module=vald
2021-11-25T01:25:54Z INF Successfully connected to EVM bridge for chain Aurora module=vald`}
</CodeBlock>

## Register as a maintainer of external chains

For each external chain C you selected earlier you must inform the Axelar network of your intent to maintain C. This is accomplished via the `register-chain-maintainer` command.
Since `vald` uses the broadcaster account,
it is recommended to be shut down when submitting this command to avoid
running into sequence mismatch errors due to concurrent signing.

Example:

```bash
axelard tx nexus register-chain-maintainer avalanche ethereum fantom moonbeam polygon binance aurora --from broadcaster --chain-id $AXELARD_CHAIN_ID --home $AXELARD_HOME --gas auto --gas-adjustment 1.4
```

<Callout type="warning" emoji="⚠️">
  Automatic deregistration

Your validator could be automatically deregistered as a maintainer for chain C for poor performance. See [Automatic deregistration](#automatic-deregistration) below.
</Callout>

## Deregister as chain maintainer from an external chain

If for some reason you need to deregister an external chain as a maintainer you must inform the Axelar network of every chain you intent to leave.
This is accomplished via the `deregister-chain-maintainer` command.

Example: Deregister the Avalanche chain:

```bash
axelard tx nexus deregister-chain-maintainer avalanche --from broadcaster --chain-id $AXELARD_CHAIN_ID --home $AXELARD_HOME --gas auto --gas-adjustment 1.4
```

<Callout type="warning" emoji="⚠️">
  Caution: You should also disable the RPC endpoint for C (set `start-with-bridge = false` in your `config.toml` file) and then restart vald.

</Callout>

## Always configure RPC and chain registration together

<Callout type="warning" emoji="⚠️">

To start/resume support for external chain C, always do these steps together:

- Enable RPC endpoint for C: set `start-with-bridge = true` in your `config.toml` file and restart vald.
- Register as a maintainer for C on Axelar network: `axelard tx nexus register-chain-maintainer`.

Conversely, to stop/pause support for external chain C:

- Deregister as a maintainer for C on Axelar network: `axelard tx nexus deregister-chain-maintainer`.
- Disable RPC endpoint for C: set `start-with-bridge = false` in your `config.toml` file and restart vald.

</Callout>

Why? If your RPC endpoint for C is enabled but you are not registered as a maintainer for C then your validator will post vote transactions for C but those transactions will be ignored by the Axelar network. Consequences:

- Your broadcaster account will lose funds because the Axelar network does not refund transaction fees for vote messages unless you are a registered maintainer for chain C.
- You will see spurious error messages in your vald logs.
- Axelar dashboards might display incorrect data on your votes for chain C.

Conversely, if you are registered as a maintainer for C but your RPC endpoint for C is disabled then your validator will fail to post vote transactions for C when the Axelar network expects them. Consequences:

- Your validator will exhibit poor vote performance and cannot earn rewards for maintaining C.

## Automatic deregistration

The Axelar network will automatically deregister your validator as a maintainer for chain C if either of the following conditions is met in the previous 500 polls for C:

1. Your validator missed at least 20% of polls. A poll is "missed" if your validator never submits a vote transaction or if the vote transaction is posted more than 3 blocks after the poll concludes.
2. Your validator voted incorrectly in at least 5% of polls. A vote is "incorrect" if it conflicts with the majority for that poll.

### How to recognize automatic deregistration

- The Axelar network emits an event each time a chain maintainer is deregistered.
- `axelar-core` logs will contain a message of the form `deregistered validator {address} as maintainer for chain {chain}`.
- Run the CLI query `axelard q nexus chain-maintainers [chain]` to see whether your validator is included.
