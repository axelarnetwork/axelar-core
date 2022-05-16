# Controller operations

import Callout from 'nextra-theme-docs/callout'

Learn how to:

- [Add a new EVM chain to the Axelar network](/controller/add-evm-chain)
- [Add a new Cosmos token or ERC-20 token to Axelar network](/controller/deploy-token)
- [coming soon] Initiate keygen and key rotation among validators for a EVM chain

A _controller_ is a special Axelar account with privileges to execute certain `axelard` CLI commands for the above tasks.

<Callout emoji="📝">
  Most participants do not need this information

Most Axelar roles (end user, node operator, validator, etc) do not need the information in this article. Many CLI commands in this article can be executed only from a controller account on Axelar network.
</Callout>

## Prerequisites

- Your fully-synced Axelar node has an account you control named `controller`. You might also have an account named `validator`. All accounts have enough AXL tokens to pay relayer gas fees for Axelar network.
- Your `controller` account is a registered controller on the Axelar network.
