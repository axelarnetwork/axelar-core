# Axelar command-line interface (CLI)

import Callout from 'nextra-theme-docs/callout'

Some CLI commands require access to a fully synced Axelar node. Learn how to [start your own Axelar node](../node/join).

Some CLI commands require AXL tokens to pay for on-chain transaction fees. Get testnet AXL tokens from the Axelar faucets: [testnet](https://faucet.testnet.axelar.dev/) | [testnet-2](https://faucet-casablanca.testnet.axelar.dev/).

Use the Axelar CLI to execute cross-chain token transfers:

- [Send UST to an EVM chain](./cli/ust-to-evm)
- [Redeem UST from an EVM chain](./cli/ust-from-evm)
- [Send AXL to an EVM chain](./cli/axl-to-evm)
- [Redeem AXL from an EVM chain](./cli/axl-from-evm)

In addition to the Axelar-specific CLI features mentioned above, Axelar also offers the same basic set of CLI commands as any other Cosmos SDK project.

[Complete Axelar CLI reference](https://github.com/axelarnetwork/axelar-core/tree/main/docs/cli)

<Callout emoji="💡">
  Tip: If you submit a transaction and encounter a out of gas error, use the following flags to set the gas manually.

```bash
--gas=auto --gas-adjustment=1.5
```

</Callout>
