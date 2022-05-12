# Convert between ETH and WETH

import Callout from 'nextra-theme-docs/callout'

Axelar supports cross-chain transfer of Wrapped Ether (WETH) token from the Ethereum mainnet. WETH is an ERC-20 version of Ether (ETH), Ethereum's native token.

If you have ETH tokens but not WETH tokens then you can convert your ETH to WETH for use with Axelar. (Conversely, you can convert your WETH back to ETH any time you choose.)

## Prerequisites

- A Metamask account with some ETH tokens or WETH tokens.
- If you haven't already, import the WETH ERC-20 token to your Metamask account in the Ethereum network as described in [Set up Metamask](metamask).

## Connect Metamask to Etherscan

Visit the WETH ERC-20 token contract on etherscan:

- [Ethereum Ropsten testnet](https://ropsten.etherscan.io/address/0xc778417e063141139fce010982780140aa0cd5ab#writeContract)
- [Ethereum mainnet](https://etherscan.io/address/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2#writeContract)

Click the "contract" tab, then click "write contract". (The above links should take you directly to "write contract".)

Click "connect to web3" to connect your Metamask account.

## Convert ETH to WETH

In the "deposit" field enter the amount of ETH you wish to convert to WETH.

![WETH conversion screenshot](/images/weth-etherscan.png)

Click "write". Your Metamask wallet should appear---approve the transaction and wait for the transaction to get confirmed into the Ethereum blockchain. Check your Metamask balances for the new WETH tokens.

## Convert WETH to ETH

In the "withdraw" field enter the amount of WETH (denominated in Wei) you wish to convert to ETH.

<Callout emoji="💡">
For the "withdraw" field (to convert WETH to ETH) the amount of WETH is denominated in Wei where 1 WETH = 10^18 Wei.  Example: to convert `0.2` WETH to ETH enter `200000000000000000`.

By contrast, for the "deposit" field (to convert ETH to WETH) the amount of ETH is denominated in ETH. Example: to convert `0.2` ETH to WETH enter `0.2`.
</Callout>

As above, click "write", approve the transaction, and check your Metamask for the new ETH tokens.
