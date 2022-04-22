# Get started

import Button from '../../components/button'

## Develop and test in a local development environment

The Axelar local development environment emulates multiple EVM chains and the Axelar overlay network that connects them.

1. Create new emulated EVM chains pre-loaded with ERC-20 tokens and gateway contracts.
2. Write your own `IAxelarExecutable` contracts and deploy to your emulated EVM chains.
3. Call your `IAxelarExecutable` contracts from any chain via that chain's gateway contract. Use `relay()` to simulate the Axelar overlay network.

Learn more at the `axelar-local-dev` github README:

<Button title="Axelar local development environment" url="https://github.com/axelarnetwork/axelar-local-dev" />

## Deploy to testnet or mainnet

When you're ready to go live:

1. Deploy your `IAxelarExecutable` contracts to any EVM chain supported by Axelar.
2. Remove calls to `relay()`---the Axelar network will handle everything for you!

See `axelar-local-gmp-examples` for a working example:

<Button title="axelar-local-gmp-examples" url="https://github.com/axelarnetwork/axelar-local-gmp-examples" />

## More examples

- [Simple](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/general-message-passing). Set up two EVM chains, transfer tokens from one chain to the other, send a "Hello world!" message to a contract on both chains.
- [Metamask](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/metamask). Set up two EVM chains and a simple web page to connect Metamask and transfer tokens from one chain to the other.
- [Remote](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/remote). Set up a test environment and connect to it remotely.
- [Token linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/token-linker). Use cross-chain contract calls to transfer ERC-20 tokens across EVM chains.
- [NFT linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/nft-linker). Use cross-chain contract calls to transfer NFTs across EVM chains.
- More to come
