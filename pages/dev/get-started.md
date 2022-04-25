# Get started

import Button from '../../components/button'

## Develop your cross-chain dapp in 2 simple steps

1. **_Build._** Develop your dapp, test against the Axelar local development environment.
2. **_Deploy._** Deploy your contracts, point your dapp to a live network: testnet or mainnet.

## Example: Hello-world

The `testnet` project is a complete, working example illustrating the build-deploy process.

1. Run the dapp in the local development environment.
2. Run the same dapp in the live testnet, interacting with contracts already deployed on Ethereum and Avalanche testnets.

View the github README for instructions and code:

<Button title="testnet example" url="https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/testnet" />

## Build

Build your cross-chain dapp from basic components:

- **RPC endpoints** to query or post transactions to the various EVM chains supported by Axelar.
- **Contract addresses** on various EVM chains for
  - Axelar services such as the Gateway contract and ERC-20 token contracts.
  - Your own custom `IAxelarExecutable` smart contracts.

The _Axelar local development environment_ emulates multiple EVM chains and the Axelar overlay network that connects them.

1. Create new emulated EVM chains with RPC endpoints on your localhost. These chains come pre-loaded with Gateway and ERC-20 tokens contracts.
2. Deploy your custom `IAxelarExecutable` contracts to your emulated EVM chains.
3. Test your app against the RPC endoints and contract addresses of your local development environment.

## Deploy to testnet or mainnet

When you're ready to go live:

- Deploy your custom `IAxelarExecutable` contracts to the live EVM chains your dapp supports.
- Swap out the RPC endpoints and contract addresses so they now point to live EVM chains.

Supported EVM Networks:

- Ethereum
- Avalanche
- Fantom
- Polygon
- Moonbeam
 
Any capitalization is supported as the input but incomming values for `sourceChain` will match those seen above.

## More examples

- [Simple](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/general-message-passing). Set up two EVM chains, transfer tokens from one chain to the other, send a "Hello world!" message to a contract on both chains.
- [Metamask](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/metamask). Set up two EVM chains and a simple web page to connect Metamask and transfer tokens from one chain to the other.
- [Remote](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/remote). Set up a test environment and connect to it remotely.
- [Token linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/token-linker). Use cross-chain contract calls to transfer ERC-20 tokens across EVM chains.
- [NFT linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/advanced-examples/nft-linker). Use cross-chain contract calls to transfer NFTs across EVM chains.
- More to come
