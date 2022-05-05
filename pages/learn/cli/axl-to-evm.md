# Send AXL to an EVM chain

import Callout from 'nextra-theme-docs/callout'

Transfer AXL tokens from Axelar to an EVM chain using the terminal.

## Prerequisites

- Skill level: intermediate
- You have downloaded the Axelar blockchain and are comfortable with [Basic node management](/node/basic).
- Your Axelar node has an account named `my_account` that you control. Let `{MY_ADDR}` denote the address of your `my_account` account.
- Select an EVM chain `{EVM_CHAIN}` from: Ethereum, Avalanche, Fantom, Moonbeam, Polygon.
- Complete steps from [Metamask for EVM chains](/resources/metamask) to connect your Metamask to `{EVM_CHAIN}`.
- You need both AXL tokens and `{EVM_CHAIN}` tokens to pay transaction fees.
  - **Testnet:**
    - Get some `{EVM_CHAIN}` testnet tokens as per [Metamask for EVM chains](/resources/metamask).
    - Get some AXL tokens from the Axelar faucets: [testnet](https://faucet.testnet.axelar.dev/) | [testnet-2](https://faucet-casablanca.testnet.axelar.dev/).
  - **Mainnet:** You are responsible for obtaining your own tokens.
- `{EVM_DEST_ADDR}` is an address controlled by you on the external EVM chain `{EVM_CHAIN}`. (In your Metamask, for example.) This is where your AXL tokens will be sent.
- `{AMOUNT}` is the amount of AXL tokens you wish to transfer, denominated in `uaxl`. Recall that `1 AXL = 1000000 uaxl`. See [Testnet resources](/resources/testnet) or [Mainnet resources](/resources/mainnet) for minimum transfer amounts.

## Send AXL tokens from Axelar to an EVM chain

Optional: Verify that your `my_account` account has sufficient balance as per [Basic node management](/node/basic).

Link your `{EVM_DEST_ADDR}` to a new temporary deposit address on Axelar:

```bash
axelard tx axelarnet link {EVM_CHAIN} {EVM_DEST_ADDR} uaxl --from my_account
```

Output should contain

```
successfully linked {AXELAR_TEMP_ADDR} and {EVM_DEST_ADDR}
```

Optional: query your new `{AXELAR_TEMP_ADDR}`:

```bash
axelard q nexus latest-deposit-address axelarnet {EVM_CHAIN} {EVM_DEST_ADDR}
```

Send `{AMOUNT}` of `uaxl` to the new `{AXELAR_TEMP_ADDR}`.

```bash
axelard tx bank send my_account {AXELAR_TEMP_ADDR} {AMOUNT}uaxl --from my_account
```

<Callout emoji="📝">
  Note: Third-party monitoring tools will automatically complete the remaining steps of this process.

Wait a few minutes then check your Metamask for the AXL tokens. Don't forget to import the AXL token into Metamask so you can see your balance as described in [Metamask for EVM chains](/resources/metamask).
</Callout>

<Callout type="warning" emoji="⚠️">
  Caution: If you attempt the remaining steps while third-party monitoring tools are active then your commands are likely to conflict with third-party commands. In this case you are likely to observe errors. Deeper investigation might be needed to resolve conflicts and complete the transfer.

The remaining steps are needed only if there are no active third-party monitoring tools and you wish to complete the process manually.
</Callout>

Confirm the deposit transaction. Look for `{TX_HASH}` in the output of the previous command.

```bash
axelard tx axelarnet confirm-deposit {TX_HASH} {AMOUNT}uaxl {AXELAR_TEMP_ADDR} --from my_account
```

Create and sign pending transfers for `{EVM_CHAIN}`.

```bash
axelard tx evm create-pending-transfers {EVM_CHAIN} --from my_account
axelard tx evm sign-commands {EVM_CHAIN} --from my_account
```

Output should contain

```
successfully started signing batched commands with ID {BATCH_ID}
```

Get the `execute_data`:

```bash
axelard q evm batched-commands {EVM_CHAIN} {BATCH_ID}
```

Wait for `status: BATCHED_COMMANDS_STATUS_SIGNED` and copy the `execute_data`.

Use Metamask to send a transaction on `{EVM_CHAIN}` with the `execute_data` to the Axelar gateway contract address `{GATEWAY_ADDR}`.

<Callout type="error" emoji="🔥">
  Danger: Post your transaction to the correct chain! Set your Metamask network to `{EVM_CHAIN}`.
</Callout>

<Callout type="warning" emoji="⚠️">
  Caution: Manually increase the gas limit to 5 million gas (5000000). If you don't do this then the transaction will fail due to insufficient gas and you will not receive your tokens.

Before you click "confirm": select "EDIT", change "Gas Limit" to 5000000, and "Save"
</Callout>

<Callout emoji="💡">
  Tip: Learn the Axelar `{GATEWAY_ADDR}` for `{EVM_CHAIN}` in two ways:

### 1. Documentation

[Testnet resources](/resources/testnet), [Mainnet resources](/resources/mainnet).

### 2. Terminal

```bash
axelard q evm gateway-address {EVM_CHAIN}
```

</Callout>

To send a transaction to `{GATEWAY_ADDR}` using Metamask: paste hex from `execute_data` above into "Hex Data" field. (Do not send tokens!)

You should see `{AMOUNT}` of asset AXL in your `{EVM_CHAIN}` Metamask account.

Congratulations! You have transferred AXL tokens from Axelar to an external EVM chain!
