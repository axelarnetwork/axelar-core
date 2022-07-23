# Intro to Axelar for developers

Axelar enables secure, any-to-any communication across blockchains, regardless of consensus mechanism or message payload.

Cross-chain apps built on Axelar are truly permissionless, meaning that their transactions cannot be censored by any oracle, relayer or validator. Axelar is based on the same proof-of-stake security model as many of the chains it connects.

## Basic functionality

Here are two basic cross-chain functions you can add to a dApp using Axelar.

1. Token transfers: Send & receive fungible tokens securely from any chain to any chain, including Cosmos-to-EVM and other complex transfers.
2. General Message Passing: Call any function on any EVM chain from inside a dApp; compose DeFi functions; move NFTs cross-chain; perform cross-chain calls of any kind that sync state securely between dApps on various ecosystems.

When executing token transfers or calling functions on remote chains, you need to specify a `destinationChain` as a string. See [chain names](./build/chain-names) for a list of valid supported values for this string.

The following sections will show you how to set up a local development environment for testing this functionality in your dApp, then deploy to testnet and mainnet.
