# Get started

import Button from '../../components/button'

## Summary: Develop a cross-chain dApp in 2 simple steps

The ideal development process is completed in two steps: 

1. Build. Develop your dApp and test it against the Axelar local development environment.
2. Deploy. Deploy your contracts and point your dApp to a live network: testnet or mainnet.

**_To begin, download the `axelar-local-gmp-examples` repository, which contains a full suite of starter templates that are run against the Axelar local development environment._**

View the github README for instructions and code:

<Button title="Starter templates" url="https://github.com/axelarnetwork/axelar-local-gmp-examples" />

## Build

**_Build a cross-chain dApp in the local development environment using two basic components:_**

1. RPC endpoints to query or post transactions to the various EVM chains supported by Axelar.
2. Contract addresses on various EVM chains for:
    - Axelar services such as the Gateway contract and ERC-20 token contracts.
    - Your own custom `IAxelarExecutable` smart contracts.

**_The Axelar local development environment allows you to:_**

1. Create simulated EVM chains with RPC endpoints on your localhost. These chains come pre-loaded with the AxelarGateway, AxelarGasReceiver and a routed ERC-20 token contract (axlUSDC).
2. Deploy your custom `IAxelarExecutable` contracts to your simulated EVM chains.
3. Test your app against the RPC endoints and contract addresses of your local development environment.

## Deploy

When you're ready to go live to testnet or mainnet: 

1. Deploy your custom `IAxelarExecutable`contracts to the live EVM chains (testnet or mainnet) your dApp supports. 
2. Swap out the RPC endpoints and contract addresses so they now point to live EVM chains (testnet or mainnet).

## Guided video walkthroughs

### Build, test, & deploy in this three part end-to-end demo
1. Set up local environment ( Part 1/3, ~8 minutes) [video](https://www.youtube.com/watch?v=PWXmsP_a-ck)
<iframe width="560" height="315" src="https://www.youtube.com/embed/PWXmsP_a-ck" title="YouTube video player" frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>

2. Deploy and run examples locally ( Part 2/3, ~9 minutes) [video](https://www.youtube.com/watch?v=l2MAZKEWzZ4)
<iframe width="560" height="315" src="https://www.youtube.com/embed/l2MAZKEWzZ4" title="YouTube video player" frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>

3. Deploy and run examples in testnets ( Part 3/3, ~7 minutes) [video](https://www.youtube.com/watch?v=X6HwmL6Tbg0)
<iframe width="560" height="315" src="https://www.youtube.com/embed/X6HwmL6Tbg0" title="YouTube video player" frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>

### Example: NFT linker

The [axelar-local-gmp-examples](https://github.com/axelarnetwork/axelar-local-gmp-examples) repo contains an example [nft-linker](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/nft-linker) on cross-chain transfer for ERC-721 NFT tokens.

See the accompanying [video](https://www.youtube.com/watch?v=pAxuQ7PIl8g):

<iframe width="560" height="315" src="https://www.youtube.com/embed/pAxuQ7PIl8g" title="YouTube video player" frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>
