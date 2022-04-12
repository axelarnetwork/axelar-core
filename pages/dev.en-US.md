# Developer

import Button from '../components/button'

## TEST: Develop your cross-chain dapp in 2 simple steps

1. **_Build._** Develop and test in the Axelar local development environment
2. **_Deploy._** Deploy to a live network: testnet or mainnet

## Example: Hello-world

`axelar-local-dev-sample` is a complete, working example illustrating the build-deploy process.

1. Run the dapp in the local development environment.
2. Run the same dapp in the live testnet, interacting with contracts already deployed on Ethereum and Avalanche testnets.

View the github README for instructions and code:

<Button title="axelar-local-dev-sample" url="https://github.com/axelarnetwork/axelar-local-dev-sample" />

## Build

The Axelar local development environment emulates multiple EVM chains and the Axelar overlay network that connects them.

1. Create new emulated EVM chains pre-loaded with ERC-20 tokens and gateway contracts.
2. Write your own `IAxelarExecutable` contracts and deploy to your emulated EVM chains.
3. Call your `IAxelarExecutable` contracts from any chain via that chain's gateway contract. Use `relay()` to simulate the Axelar overlay network.

Learn more at the `axelar-local-dev` github README:

<Button title="Axelar local development environment" url="https://github.com/axelarnetwork/axelar-local-dev" />

## Deploy

When you're ready to go live:

1. Deploy your `IAxelarExecutable` contracts to any EVM chain supported by Axelar.
2. Remove calls to `relay()`---the Axelar network will handle everything for you!

See `axelar-local-dev-sample` for a working example:

<Button title="axelar-local-dev-sample" url="https://github.com/axelarnetwork/axelar-local-dev-sample" />

## More examples

- [Simple](https://github.com/axelarnetwork/axelar-local-dev/tree/main/examples/simple). Set up two EVM chains, transfer tokens from one chain to the other, send a "Hello world!" message to a contract on both chains.
- [Metamask](https://github.com/axelarnetwork/axelar-local-dev/tree/main/examples/metamask). Set up two EVM chains and a simple web page to connect Metamask and transfer tokens from one chain to the other.
- [Remote](https://github.com/axelarnetwork/axelar-local-dev/tree/main/examples/remote). Set up a test environment and connect to it remotely.
- [Token linker](https://github.com/axelarnetwork/axelar-local-dev/tree/main/examples/tokenLinker). Use cross-chain contract calls to transfer ERC-20 tokens across EVM chains.
- More to come

## AxelarJS SDK

The AxelarJS SDK is a `npm` dependency that empowers developers to leverage microservices or IBC relayers provided by Axelar.

Not all dapps need the SDK. Your dapp might benefit from the SDK in the following use cases:

1. **_Microservices._** Example: Get a deposit address for cross-chain token transfer dapps like [Satellite](/resources/satellite).
2. **_Relayers._** Example: Connect EVM chains to Cosmos chains such as Terra.

Learn more at the `axelarjs-sdk` github README:

<Button title="axelarjs-sdk" url="https://github.com/axelarnetwork/axelarjs-sdk" />
