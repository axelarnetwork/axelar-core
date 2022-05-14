# Mainnet

import MarkdownPath from '../../components/markdown'
import EVMChains from '../../components/evm/chains'
import EVMAssets from '../../components/evm/assets'
import IBCChannels from '../../components/ibc/channels'

| Variable              | Value     |
| --------------------- | --------- |
| `axelar-core` version | `v0.17.1` |
| `vald` version        | `v0.17.0` |
| `tofnd` version       | `v0.10.1`  |

<div className="space-y-1 mt-4">
  ## EVM Chains
  <EVMChains environment="mainnet" />
</div>

<div className="space-y-1 mt-4">
  ## Assets
  <EVMAssets environment="mainnet" />
</div>

<div className="space-y-1 mt-4">
  ## IBC Channels
  <IBCChannels environment="mainnet" />
</div>

## Cross-chain transfer fee

The Network (and thus the Satellite app) charges a base fee for all cross-chain transfers.
Most up-to-date values are present in the Satellite app.
This fee only depends on the source/destination chain and the asset and does NOT take a percentage from the transfer amount.
When transferring an asset X from chain Y to chain Z, the transfer fee is the sum of per-chain fee for that asset.
For e.g. a transfer of 1000 USDC from Ethereum to Osmosis will have a fee of 40.5 USDC (= 40 USDC for Ethereum + 0.5 USDC for Osmosis),
and so the recipient will get 959.5 USDC.

| Asset symbol | Ethereum   | non-Ethereum EVM | Cosmos Chains | Decimals | Unit         |
| ------------ | ---------- | ---------------- | ------------- | -------- | ------------ |
| USDC         | 40 USDC    | 1 USDC           | 0.5 USDC      | 6        | uusdc        |
| WETH         | 0.02 WETH  | N/A              | 0.0002 WETH   | 18       | weth-wei     |
| WBTC         | 0.002 WBTC | N/A              | 0.00002 WBTC  | 8        | wbtc-satoshi |
| DAI          | 40 DAI     | 1 DAI            | 0.5 DAI       | 18       | dai-wei      |
| FRAX         | 40 FRAX    | 1 FRAX           | 0.5 FRAX      | 18       | frax-wei     |
| USDT         | 40 USDT    | 1 USDT           | 0.5 USDT      | 6        | uusdt        |
| ATOM         | 4 ATOM     | 0.1 ATOM         | 0.05 ATOM     | 6        | uatom        |
| UST          | 80 UST     | 2 UST            | 1 UST         | 6        | uusd         |
| LUNA         | 40 LUNA    | 1 LUNA           | 0.5 LUNA      | 6        | uluna        |
| NGM          | 40 NGM     | 1 NGM            | 0.5 NGM       | 6        | ungm         |
| EEUR         | 40 EEUR    | 1 EEUR           | 0.5 EEUR      | 6        | eeur         |

The current transfer fee can also be queried on the network with

```bash
axelard q nexus transfer-fee [source chain] [destination chain] [amount]
```

For e.g., querying the example transfer above (note `1 UST = 10^6 uusd`),

```bash
axelard q nexus transfer-fee terra avalanche 1000000000uusd
```

The per-chain fee info can be queried via

```bash
axelard q nexus fee avalanche uusd
```

If the total amount of asset X sent to a deposit address A is NOT greater than the transfer fee,
then those deposits will sit in the queue until a future deposit to A brings the total above the fee.

Additionally, users should be prepared to pay for any transaction fees assessed by the source chain when transferring funds into a deposit address.
These fees are typically in the form of native tokens on that chain (for e.g. LUNA on Terra, ETH on Ethereum).

## Upgrade Path

<MarkdownPath src="/md/mainnet/upgrade-path.md" />
