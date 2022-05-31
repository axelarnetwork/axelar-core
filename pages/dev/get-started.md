# Get started

import Button from '../../components/button'

## Develop your cross-chain dapp in 2 simple steps

1. **_Build._** Develop your dapp, test against the Axelar local development environment.
2. **_Deploy._** Deploy your contracts, point your dapp to a live network: testnet or mainnet.

## Examples

### Examples repo

There are several complete, working examples with instructions at the `axelar-local-gmp-examples` repo that illustrate the build-deploy process.

View the github README for instructions and code:

<Button title="examples" url="https://github.com/axelarnetwork/axelar-local-gmp-examples" />

### Video on NFT linker example

The [axelar-local-gmp-examples](https://github.com/axelarnetwork/axelar-local-gmp-examples) repo contains an example [nft-linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/nft-linker) on cross-chain transfer for ERC-721 NFT tokens.

See the accompanying [video](https://www.youtube.com/watch?v=pAxuQ7PIl8g):

<iframe width="560" height="315" src="https://www.youtube.com/embed/pAxuQ7PIl8g" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

## Build

Build your cross-chain dapp from basic components:

- **RPC endpoints** to query or post transactions to the various EVM chains supported by Axelar.
- **Contract addresses** on various EVM chains for
  - Axelar services such as the Gateway contract and ERC-20 token contracts.
  - Your own custom `IAxelarExecutable` smart contracts.

The _Axelar local development environment_ emulates multiple EVM chains and the Axelar overlay network that connects them.

1. Create new emulated EVM chains with RPC endpoints on your localhost. These chains come pre-loaded with the AxelarGateway, AxelarGasReceiver and a routed ERC-20 token contract (UST).
2. Deploy your custom `IAxelarExecutable` contracts to your emulated EVM chains.
3. Test your app against the RPC endoints and contract addresses of your local development environment.

## Deploy to testnet or mainnet

When you're ready to go live:

- Deploy your custom `IAxelarExecutable` contracts to the live EVM chains your dapp supports.
- Swap out the RPC endpoints and contract addresses so they now point to live EVM chains.
