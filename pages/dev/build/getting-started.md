import DevNav from '../../../components/index/dev-nav'

# Building on Axelar

Now that you've run the simple "Hello World" example, this "Build" section will walk you through the suite of Axelar tools you can use to build and deploy your own cross-chain dApp. 

The ideal development process is completed in two steps: 

1. Build. Develop your dApp and test it against the Axelar local development environment.
2. Deploy. Deploy your contracts and point your dApp to a live network: testnet or mainnet.

## Build
**A cross-chain dApp on Axelar consists of two components:**

1. RPC endpoints to query or post transactions to the various EVM chains supported by Axelar.
2. Contract addresses on various EVM chains for:
    - Axelar services such as the Gateway contract and ERC-20 token contracts.
    - Your own custom `IAxelarExecutable` smart contracts.

**The Axelar local development environment allows you to:**

* Create simulated EVM chains with RPC endpoints on your localhost. These chains come preloaded with the AxelarGateway, AxelarGasReceiver and a routed ERC-20 token contract (axlUSDC).
* Deploy your custom `IAxelarExecutable` contracts to your simulated EVM chains.
* Test your dApp against the RPC endoints and contract addresses of your local development environment.


For the actual development of your dApp, there are four simple steps (two mandatory and two optional):
1. Call the contract method on the source chain:
    - `callContract`, or
    - `callContractWithToken`
2. Pay gas to the Gas Services contract on the source chain.
3. Optional: Check the status of the call.
4. Optional: Execute and recover transactions.

## Deploy

When you're ready to go live to testnet or mainnet: 

1. Deploy your custom `IAxelarExecutable`contracts to the live EVM chains (testnet or mainnet) your dApp supports. 
2. Swap out the RPC endpoints and contract addresses so they now point to live EVM chains (testnet or mainnet).

## Helpful resources
If you have any issues with the two steps above, you can use the suite of tools and other "kick-starter" examples in the Axelar developer ecosystem to get you going:

<br/>
<DevNav />