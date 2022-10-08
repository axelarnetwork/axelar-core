# Convert between native and wrapped tokens

import Callout from 'nextra-theme-docs/callout'

Axelar supports cross-chain transfer of the following native tokens via their wrapped ERC-20 versions:

- AVAX (Avalanche)
- ETH (Ethereum)
- FTM (Fantom)
- GLMR (Moonbeam)
- MATIC (Polygon)

If you have native tokens but not wrapped tokens then you can convert your native to wrapped for use with Axelar. (Conversely, you can convert your wrapped back to native any time you choose.)

## Prerequisites

A Metamask account with some native tokens or wrapped tokens.

## Connect Metamask to a block explorer

Visit the wrapped ERC-20 token contract on the appropriate block explorer:

Testnets:

- [Avalanche Fuji testnet](https://testnet.snowtrace.io/token/0xd00ae08403B9bbb9124bB305C09058E32C39A48c#writeContract)
- [Ethereum Goerli testnet](https://goerli.etherscan.io/address/0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6#writeContract)
- [Fantom testnet](https://testnet.ftmscan.com/token/0x812666209b90344Ec8e528375298ab9045c2Bd08#writeContract)
- [Moonbase Alpha testnet](https://moonbase.moonscan.io/address/0x1436aE0dF0A8663F18c0Ec51d7e2E46591730715#writeContract)
- [Polygon Mumbai testnet](https://mumbai.polygonscan.com/address/0x9c3c9283d3e44854697cd22d3faa240cfb032889#writeContract)

Mainnets:

- [Avalanche](https://snowtrace.io/token/0xb31f66aa3c1e785363f0875a1b74e27b85fd66c7#writeContract)
- [Ethereum](https://etherscan.io/address/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2#writeContract)
- [Fantom](https://ftmscan.com/token/0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83#writeContract)
- [Moonbeam](https://moonbeam.moonscan.io/token/0xacc15dc74880c9944775448304b263d191c6077f#writeContract)
- [Polygon](https://polygonscan.com/token/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270#writeContract)

If you haven't already, import the wrapped ERC-20 token to your Metamask account in the appropriate network as described in [Set up Metamask](metamask).

Click the "contract" tab, then click "write contract". (The above links should take you directly to "write contract".)

Click "connect to web3" to connect your Metamask account.

## Convert native to wrapped

In the "deposit" field enter the amount of native token you wish to convert to wrapped token. (The following screenshot is for Ethereum.)

![WETH conversion screenshot](/images/weth-etherscan.png)

Click "write". Your Metamask wallet should appear---approve the transaction and wait for the transaction to get confirmed into the blockchain. Check your Metamask balances for the new wrapped tokens.

## Convert wrapped to native

In the "withdraw" field enter the amount of wrapped tokens (denominated in Wei) you wish to convert to native.

<Callout emoji="💡">
For the "withdraw" field (to convert wrapped to native) the amount of wrapped tokens is denominated in Wei where 1 wrapped token = 10^18 Wei.  Example: to convert `0.2` WETH to ETH enter `200000000000000000`.

By contrast, for the "deposit" field (to convert native to wrapped tokens) the amount of native is denominated in native units. Example: to convert `0.2` ETH to WETH enter `0.2`.
</Callout>

As above, click "write", approve the transaction, and check your Metamask for the new native tokens.
