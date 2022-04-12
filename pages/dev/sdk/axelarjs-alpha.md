# Alpha (v0.5.xx)

## What you can do with v0.5.xx

1. Get a deposit address for cross-chain token transfer
2. **New!** Cross-chain EVM contract calls

## Install

Install the latest patch of AxelarJS SDK v0.5.xx:

```bash
npm i --save @axelar-network/axelarjs-sdk@alpha
```

## Get a deposit address for cross-chain token transfer

For cross-chain token transfers from a source chain `X` to destination chain `Y` for asset `Z`, Axelar generates a deposit address on the source chain `X` that accepts deposits of asset `Z` that are then relayed through a collection of microservices through the Axelar network to the destination chain `Y`. 

Below is a sample function `myGetDepositAddress` that returns a new deposit address `A` on the Axelar chain. The function `myGetDepositAddress` wraps a call to `getDepositAddress` from the AxelarJS SDK API like so:

```typescript
import {
  AxelarAssetTransfer,
  Environment
} from "@axelar-network/axelarjs-sdk"

const api = new AxelarAssetTransfer({ environment: Environment.TESTNET });

const myGetDepositAddress = async (destinationAddress?: string) => {
  const linkAddress: string = await api.getDepositAddress("axelar", "avalanche", "0x74Ccd7d9F1F40417C6F7fD1151429a2c44c34e6d", "uaxl");
  return linkAddress;
};
```

See [Deposit address demo (alpha)](./deposit-address-demo-alpha) for a working demo in the browser.

## New! Cross-chain EVM contract calls

The Axelar network supports arbitrary cross-chain contract calls for EVM chains. This feature can be used to transfer ERC-20 tokens across EVM chains---a simple and efficient alternative to deposit addresses for EVM-to-EVM token transfers.

The `AxelarGateway` library in this SDK provides Typescript-wrapped utility functions for a handful of the core methods, or you can access the contract directly.

Below is an example of how you can instantiate and invoke methods on the `AxelarGateway` SDK. 
### contractCallWithToken method

`contractCallWithToken` is invoked on the Axelar gateway contract on an EVM source chain to call a custom smart contract on a destination chain, sending tokens along with it. 

The code snippet below is an example in `devnet` that calls a custom smart contract on Ethereum from Avalanche with Axelar-wrapped UST. Alternatively, see this [repo](https://github.com/axelarnetwork/axelar-gateway-sample/blob/main/src/call-contract-with-token.ts) for a full working demo of the below.

1. Instantiate the gateway SDK:

```typescript

import "dotenv/config";
import {
  AxelarGateway,
  Environment,
  EvmChain,
} from "@axelar-network/axelarjs-sdk";
import { ethers } from "ethers";

const privateKey = process.env.PRIVATE_KEY;
const provider = new ethers.providers.JsonRpcProvider("https://api.avax-test.network/ext/bc/C/rpc");
const evmWallet = new ethers.Wallet(privateKey, provider);

const UST_ADDRESS_AVALANCHE = "0x96640d770bf4a15Fb8ff7ae193F3616425B15FFE";
const AXELAR_GATEWAY_CONTRACT = "0x4ffb57aea2295d663b03810a5802ef2bc322370d";

const gateway = AxelarGateway.create(
  Environment.DEVNET,
  EvmChain.AVALANCHE,
  provider
);

```

2. Set an allowance on the gateway contract to access funds on the token contract:

```typescript

/*helper method that invokes approval on the gateway contract for an address to move UST funds*/
async function approveTransactionIfNeeded(address: string) {
  const contract = new ethers.Contract(address, erc20Abi, provider);
  const allowance: ethers.BigNumber = await contract.allowance(
    evmWallet.address,
    AXELAR_GATEWAY_CONTRACT
  );

  const approvalRequired = allowance.isZero();

  if (approvalRequired) {
    console.log("\n==== Approving UST... ====");
    const receipt = await gateway
      .createApproveTx({ tokenAddress: UST_ADDRESS_AVALANCHE })
      .then((tx) => tx.send(evmWallet))
      .then((tx) => tx.wait());
    console.log(
      "UST has been approved to gateway contract",
      receipt.transactionHash
    );
  }
}

```

3. Invoke `callContractWithToken`, checking whether funds are approved first with the helper method above:

```typescript


const getBalance = (address: string) => {
  const contract = new ethers.Contract(address, erc20Abi, provider);
  return contract.balanceOf(evmWallet.address);
}


const callContractWithToken = async () => {

  console.log("==== Your UST balance ==== ");
  const ustBalance = await getBalance(UST_ADDRESS_AVALANCHE);
  console.log(ethers.utils.formatUnits(ustBalance, 6), "UST");

  // Check UST Approval to Gateway Contract
  await approveTransactionIfNeeded(UST_ADDRESS_AVALANCHE);

  console.log("\n==== Call contract with token ====");
  const encoder = ethers.utils.defaultAbiCoder;
  const payload = encoder.encode(["address[]"], [[evmWallet.address]]);
  const amount = ethers.utils.parseUnits("10", 6).toString();

  const callContractReceipt = await gateway
    .createCallContractWithTokenTx({
      destinationChain: EvmChain.ETHEREUM,
      destinationContractAddress: "0xB628ff5b78bC8473a11299d78f2089380f4B1939",
      payload,
      amount,
      symbol: "UST",
    })
    .then((tx) => tx.send(evmWallet))
    .then((tx) => tx.wait());

  console.log(
    "Call contract with token tx:",
    `https://testnet.snowtrace.io/tx/${callContractReceipt.transactionHash}`
  );
};

callContractWithToken();


```

### Local Development

If you are developing using a local blockchain onto which you deployed a custom Axelar gateway contract, you can alternatively inject that directly. Instead of calling `AxelarGateway.create` above, you can instantiate the gateway this way:

```typescript
const myCustomGatewayContractAddress = "0x...";

const gateway = new AxelarGateway(
  myCustomGatewayContractAddress,
  ethers.provider
);

```

### Retrieve Gateway Contract

The SDK includes abstracted methods for those defined on the Axelar Gateway contract. If you prefer to fetch the gateway contract and invoke methods directly, we expose the contract this way:

```typescript
const gatewayContract = gateway.getContract();
```