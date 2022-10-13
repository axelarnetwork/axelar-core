# Testnet

import Callout from 'nextra-theme-docs/callout'
import MarkdownPath from '../../components/markdown'
import IBCChannels from '../../components/ibc/channels'
import TransferFeeCalculator from '../../components/transfer-fee/calculator'
import Typeform from '../../components/typeform'
import EVMChains from '../../components/evm/chains'
import EVMAssets from '../../components/evm/assets'

| Variable              | Value     |
| --------------------- | --------- |
| `axelar-core` version | `v0.26.0-patch` |
| `vald` version        | `v0.26.0-patch` |
| `tofnd` version       | `v0.10.1` |

## EVM contract addresses
<div className="space-y-1 mt-4">
  <EVMChains environment="testnet" />

  `Constant Address Deployer`:  `0x98b2920d53612483f91f12ed7754e51b4a77919e`
</div>

<div className="space-y-1 mt-4">
  ### Assets
  <EVMAssets environment="testnet" />
</div>

<div className="space-y-1 mt-4">
  ## IBC Channels
  <IBCChannels environment="testnet" />
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

<div className="space-y-1 mt-4">
  <TransferFeeCalculator environment="testnet" />
</div>

The current gas relayer fee is also available via node query:

```bash
axelard q nexus transfer-fee [source chain] [destination chain] [amount]
```

Example: transfer USDC from Avalanche to Osmosis. (The amount here is arbitrary---gas relayer fees do not depend on the amount. `1 USDC = 10^6 uusdc`).

```bash
axelard q nexus transfer-fee avalanche osmosis 1000000000uusdc
```

The per-chain gas relayer fee info can be queried via

```bash
axelard q nexus fee-info avalanche uusdc
```

If the total amount of a token sent to a deposit address A is NOT greater than the gas relayer fee
then those deposits will wait in the queue until a future deposit to A brings the total above the fee.

The gas relayer fee does not include any transaction fee assessed by the source chain for transferring tokens into a deposit address. These fees are usually denominated in native tokens on that chain (for e.g. AVAX on Avalanche, ETH on Ethereum).

## Upgrade Path

<MarkdownPath src="/md/testnet/upgrade-path.md" />

<Typeform />
