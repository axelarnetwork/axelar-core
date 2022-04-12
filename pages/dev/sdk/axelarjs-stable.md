# Stable (v0.4.xx)

## Quick start: simple Axelar-to-Avalanche token transfer demo

We'll write a function `myGetDepositAddress` with

- **Input:** A user-provided `destinationAddress` on Avalanche testnet.
- **Output:** A new one-time deposit address `addr` on the Axelar testnet.

From here, a user could do the following:

- Use the [Axelar testnet faucet](https://faucet.testnet.axelar.dev/) to deposit AXL tokens to `addr`. (Be sure to exceed the minimum of 10 AXL tokens!) The Axelar network will transfer these AXL tokens to Avalanche.
- After some time, verify your AXL tokens arrived on Avalanche at `destinationAddress`. Use an Avalanche testnet block explorer such as [Snowtrace](https://testnet.snowtrace.io/). View the ERC-20 [token contract for AXL tokens on Avalanche testnet](https://testnet.snowtrace.io/address/0x46cc87ea84586c03bb2109ed9b33f998d40b7623).

You can substitute (Axelar, AXL, Avalanche) for many other choices of (source chain, asset, destination chain).

See [Deposit address demo (stable)](./deposit-address-demo-stable) for a working demo in the browser.

## Install

Install the latest patch of AxelarJS SDK v0.4.xx:

```bash
npm i --save @axelar-network/axelarjs-sdk@0.4.29
```

## Get a deposit address from the Axelar network

Write a function `myGetDepositAddress` that wraps a call to `getDepositAddress` from the AxelarJS SDK API like so:


```typescript
import {
  AssetTransferObject,
  TransferAssetBridge,
} from "@axelar-network/axelarjs-sdk";

const axelarAPI = new TransferAssetBridge(process.env.REACT_APP_STAGE);

const myGetDepositAddress = async (destinationAddress?: string) => {
  const { otc, publicAddr, signature } = await promptUserToSignMessage();
  const parameters: AssetTransferObject = getParameters(
    destinationAddress || publicAddr
  ); // wherever you specify for the destination address on the destination chain
  parameters.otc = otc;
  parameters.publicAddr = publicAddr;
  parameters.signature = signature;

  const linkAddress = await axelarAPI.getDepositAddress(parameters, true);

  return linkAddress;
};
```

What's happening in the `myGetDepositAddress` wrapper?

- `REACT_APP_STAGE`: This version of the SDK requires an environment variable `REACT_APP_STAGE` set to "testnet" or "mainnet". This requirement will be eliminated in future versions.
- `promptUserToSignMessage`: This version of the SDK employs a rate-limiting strategy to protect our microservices from abuse. Each call to `getDepositAddress` must pass a signature of a random one-time code `otc`. The function `promptUserToSignMessage` gets such a signature from the user.
- `getParameters`: prepares a struct of type `AssetTransferObject` expected by `getDepositAddress`.
- The second argument to `getDepositAddress` is an optional `boolean` indicating whether error alerts should be visible in the UI.

## Implement the one-time code signer

The function `promptUserToSignMessage` can be implemented like so:

```typescript
/*below is sample implementation using ethers.js, but you can use whatever you want*/
const provider = new ethers.providers.Web3Provider(window.ethereum, "any"); //2nd param is network type
const signerAuthority = provider.getSigner();
const signerAuthorityAddress = signerAuthority.getAddress();

const getNoncedMessageToSign = async () => {
  const signerAuthorityAddress = await signerAuthority.getAddress();
  const { validationMsg, otc } = await api.getOneTimeCode(
    signerAuthorityAddress
  );
  return { validationMsg, otc };
};

const promptUserToSignMessage = async () => {
  const { validationMsg, otc } = await getNoncedMessageToSign();
  const signature = await signerAuthority.signMessage(validationMsg);

  return {
    otc,
    publicAddr: await signerAuthority.getAddress(),
    signature,
  };
};
```

## Implement parameter generation

The function `getParameters` can be implemented like so:

```typescript
import {
  ChainList,
  ChainInfo, // TODO is this right??
} from "@axelar-network/axelarjs-sdk";

const getParameters = (
  destinationAddress: string,
  sourceChainName: string = "axelar",
  destinationChainName: string = "avalanche",
  asset_common_key: string = "uaxl"
) => {
  /*
  info for sourceChainInfo and destinationChainInfo fetched from the ChainList module of the SDK. 
  * */
  const axelarChain: ChainInfo = ChainList.map(
    (chain: Chain) => chain.chainInfo
  ).find(
    (chainInfo: ChainInfo) =>
      chainInfo.chainName.toLowerCase() === sourceChainName.toLowerCase()
  ) as ChainInfo;
  const avalancheChain: ChainInfo = ChainList.map(
    (chain: Chain) => chain.chainInfo
  ).find(
    (chainInfo: ChainInfo) =>
      chainInfo.chainName.toLowerCase() === destinationChainName.toLowerCase()
  ) as ChainInfo;
  const assetObj = axelarChain.assets?.find(
    (asset: AssetInfo) => asset.common_key === asset_common_key
  ) as AssetInfo;

  let requestPayload: AssetTransferObject = {
    sourceChainInfo: terraChain,
    destinationChainInfo: avalancheChain,
    selectedSourceAsset: assetObj,
    selectedDestinationAsset: {
      ...assetObj,
      assetAddress: destinationAddress, //address on the destination chain where you want the tokens to arrive
    },
    signature: "SIGNATURE_FROM_METAMASK_SIGN",
    otc: "OTC_RECEIVED_FROM_SERVER",
    publicAddr: "SIGNER_OF_SIGNATURE",
    transactionTraceId: "YOUR_OWN_UUID", //your own UUID, helpful for tracing purposes. optional.
  };

  return requestPayload;
};
```
