# Deploy a new token

import Callout from 'nextra-theme-docs/callout'

Learn how to deploy a new Cosmos or ERC-20 token to any EVM chain supported by Axelar.

<Callout emoji="💡">
  Example: For clarity, this article deploys the following tokens to the following EVM chains:

  - Cosmos Token: UST (Terra native token)
  - ERC-20 Token: WAVAX (Avalanche native token)
  - EVM chains: Avalanche

  Substitute your own tokens and EVM chains as desired.

  Repeat these instructions for each additional EVM chain.
</Callout>

## Prerequisites

- Prerequisites for [Controller operations](../controller)
- You will deploy smart contracts to the EVM chain---you need enough native tokens to pay gas fees on that chain. Example: if deploying to Avalanche then you need AVAX tokens, etc.
- The EVM chain has been added to the network.

## Deploy and confirm ERC-20 token contracts

The initial command to register a token contract depends on whether the token being added is native to that EVM chain.

If registering a token whose native chain is different than the EVM chain being deployed to, for e.g. deploying `UST` (native to Terra) to `Avalanche`, do the following:

```bash
axelard tx evm create-deploy-token avalanche [native chain] [asset] [erc-20 token name] [erc-20 symbol] [decimals] [capacity] --from controller --gas auto --gas-adjustment 1.4

axelard tx evm create-deploy-token avalanche terra uusd "Axelar Wrapped UST" UST 6 0 --from controller --gas auto --gas-adjustment 1.4
```

If registering an ERC-20 token native to that EVM chain, for e.g. registering `WAVAX` on `Avalanche`, then it's token contract needs to be specified, and the ERC-20 token name and symbol MUST match the value in the token contract:

Retrieve the ERC-20 info from it's token contract. For `WAVAX`, it can found [here](https://snowtrace.io/address/0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7#readContract).

```bash
axelard tx evm create-deploy-token avalanche [native chain] [asset] [erc-20 token name] [erc-20 symbol] [decimals] [capacity] --address [token contract] --from controller --gas auto --gas-adjustment 1.4

axelard tx evm create-deploy-token avalanche terra wavax-wei "Wrapped AVAX" WAVAX 18 0 --address 0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7 --from controller --gas auto --gas-adjustment 1.4
```

Sign the above token deployment commands into a batch for the gateway.
This transaction does not need controller permission---you may sign it with any account, such as your node's `validator` account.

```bash
axelard tx evm sign-commands avalanche --from validator
```

Submit `execute_data` shown below to the gateway contract on Avalanche.

```bash
axelard q evm latest-batched-commands avalanche
```

```bash
axelard q evm gateway-address avalanche
```

Submitting this batched data is similar to the description in [Send AXL to an EVM chain](../dev/cli/axl-to-evm).

- Note the `[EVM_TOKEN_TX_HASH]` for the transaction to the gateway contract.

Wait until the transaction `[EVM_TOKEN_TX_HASH]` has received enough block confirmations on the EVM chain. (This number was set in the `confirmation_height` in the file `evm-chain.json` when you executed `add-chain`.)

For each token call a validator vote to confirm deployment of the ERC-20 contract.

```bash
axelard tx evm confirm-erc20-token avalanche terra uusd [EVM_TOKEN_TX_HASH] --from controller --gas auto --gas-adjustment 1.4

axelard tx evm confirm-erc20-token avalanche avalanche wavax-wei [EVM_TOKEN_TX_HASH] --from controller --gas auto --gas-adjustment 1.4
```

Optional: check your logs for messages of the form `token XXX deployment confirmation result on chain avalanche is true`.
