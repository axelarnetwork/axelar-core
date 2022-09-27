# Testnet

import Callout from 'nextra-theme-docs/callout'
import MarkdownPath from '../../components/markdown'
import EVMChains from '../../components/evm/chains'
import EVMAssets from '../../components/evm/assets'
import IBCChannels from '../../components/ibc/channels'
import Typeform from '../../components/typeform'

| Variable              | Value     |
| --------------------- | --------- |
| `axelar-core` version | `v0.17.3` |
| `vald` version        | `v0.17.3` |
| `tofnd` version       | `v0.10.1` |

<div className="space-y-1 mt-4">
  ## EVM Chains
  <EVMChains environment="testnet-2" />
</div>

<div className="space-y-1 mt-4">
  ## Assets
  <EVMAssets environment="testnet-2" />
</div>

<div className="space-y-1 mt-4">
  ## IBC Channels
  <IBCChannels environment="testnet-2" />
</div>

## Cross-chain relayer gas fee

The Axelar network charges a _relayer gas fee_ for all cross-chain token transfers in order to pay for network-level transaction ("gas") fees across chains.
The relayer gas fee amount depends only on:

- the asset
- the source chain
- the destination chain

<Callout emoji="💡">
  The relayer gas fee does NOT take a percentage from the transfer amount.
</Callout>

Example: a transfer of X USDC tokens from Ethereum to Osmosis will have a fee of 40.5 USDC (= 40 USDC for Ethereum + 0.5 USDC for Osmosis),
and so the recipient will get X - 40.5 USDC tokens on Osmosis.

| Asset symbol | Ethereum   | non-Ethereum EVM | Cosmos Chains | Decimals | Unit         |
| ------------ | ---------- | ---------------- | ------------- | -------- | ------------ |
| USDC         | 60 USDC    | 1 USDC           | 0.5 USDC      | 6        | uusdc        |
| WETH         | 0.02 WETH  | N/A              | 0.0002 WETH   | 18       | weth-wei     |
| WBTC         | 0.002 WBTC | N/A              | 0.00002 WBTC  | 8        | wbtc-satoshi |
| DAI          | 60 DAI     | 1 DAI            | 0.5 DAI       | 18       | dai-wei      |
| FRAX         | 60 FRAX    | 1 FRAX           | 0.5 FRAX      | 18       | frax-wei     |
| USDT         | 60 USDT    | 1 USDT           | 0.5 USDT      | 6        | uusdt        |
| ATOM         | 4 ATOM     | 0.1 ATOM         | 0.05 ATOM     | 6        | uatom        |
| UST          | 80 UST     | 2 UST            | 1 UST         | 6        | uusd         |
| LUNA         | 40 LUNA    | 1 LUNA           | 0.5 LUNA      | 6        | uluna        |
| NGM          | 60 NGM     | 1 NGM            | 0.5 NGM       | 6        | ungm         |
| EEUR         | 60 EEUR    | 1 EEUR           | 0.5 EEUR      | 6        | eeur         |

The current gas relayer fee is also available via node query:

```bash
axelard q nexus transfer-fee [source chain] [destination chain] [amount]
```

Example: transfer USDC from Ethereum to Osmosis. (The amount here is arbitrary---gas relayer fees do not depend on the amount. `1 USDC = 10^6 uusdc`).

```bash
axelard q nexus transfer-fee terra osmosis 1000000000uusdc
```

The per-chain gas relayer fee info can be queried via

```bash
axelard q nexus fee-info avalanche uusdc
```

If the total amount of a token sent to a deposit address A is NOT greater than the gas relayer fee
then those deposits will wait in the queue until a future deposit to A brings the total above the fee.

The gas relayer fee does not include any transaction fee assessed by the source chain for transferring tokens into a deposit address. These fees are usually denominated in native tokens on that chain (for e.g. AVAX on Avalanche, ETH on Ethereum).

## Upgrade Path

<MarkdownPath src="/md/testnet-2/upgrade-path.md" />

<Typeform />
