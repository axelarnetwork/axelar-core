# Testnet

import MarkdownPath from '../../components/markdown'
import EVMAddresses from '../../components/evm/addresses'
import IBCChannels from '../../components/ibc/channels'

| Variable              | Value         |
| ----------------------| ------------- |
| `axelar-core` version | `v0.16.2`     |
| `tofnd` version       | `v0.8.2`      |

<div className="space-y-1 mt-4">
  ## EVM Chains
  <EVMAddresses environment="testnet" />
</div>

<div className="space-y-1 mt-4">
  ## IBC Channels
  <IBCChannels environment="testnet" />
</div>

## Cross-chain transfer fee

The Network (and thus the Satellite app) charges a base fee for all cross-chain transfers.
This fee only depends on the source/destination chain and the asset and does NOT take a percentage from the transfer amount.
When transferring an asset X from chain Y to chain Z, the transfer fee is the sum of per-chain fee for that asset.
For e.g. a transfer of 1000 UST from Terra to Avalanche will have a fee of 1.5 UST (= 0.5 UST for Terra + 1.0 UST for Avalanche), and so the recipient will get 998.5 UST.

| Asset symbol | Ethereum | non-Ethereum EVM | Cosmos Chains  | Decimals  | Unit     |
| ------------ | -------- | ---------------- | -------------- | --------- | -------- |
| UST          | 20 UST   | 1 UST            | 0.5 UST        | 6         | uusd     |
| LUNA         | 0.2 LUNA | 0.01 LUNA        | 0.005 LUNA     | 6         | uluna    |
| ATOM         | 0.7 ATOM | 0.04 ATOM        | 0.02 ATOM      | 6         | uatom    |
| aUSDC        | 20 aUSDC | 1 aUSDC          | 0.5 aUSDC      | 6         | uausdc   |

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

<MarkdownPath src="/md/testnet/upgrade-path.md" />