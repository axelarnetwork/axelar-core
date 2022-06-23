## Query and Recover GMP transactions

Occasionally, transactions can get "stuck" in the pipeline from a source to destination chain (e.g. due to one-off issues that arise with relayers that operate on top of the network).

The `AxelarGMPRecoveryAPI` module in the AxelarJS SDK can be used by your dApp to query the status of any General Message Passing (GMP) transaction (triggered by either `callContract` or `callContractWithToken`) on the gateway contract of a source chain and trigger a manual relay from source to destination if necessary. - The [[GMP status tracker](../gmp-tracker)] on Axelarscan makes use of this feature.

### Install the AxelarJS SDK module (AxelarGMPRecoveryAPI)

Install the AxelarJS SDK:

```bash
npm i @axelar-network/axelarjs-sdk@alpha
```

Instantiate the `AxelarGMPRecoveryAPI` module:

```ts
import {
  AxelarGMPRecoveryAPI,
  Environment,
} from "@axelar-network/axelarjs-sdk";

const sdk = new AxelarGMPRecoveryAPI({
  environment: Environment.TESTNET,
});
```

### Query transaction status by txHash

Invoke `queryTransactionStatus`:

```ts
const txHash: string =
  "0xfb6fb85f11496ef58b088116cb611497e87e9c72ff0c9333aa21491e4cdd397a";
const txStatus: GMPStatusResponse = sdk.queryTransactionStatus(txHash);
```

where possible status responses for txStatus are outlined below:

```ts
interface GMPStatusResponse {
  status: GMPStatus;
  details: any;
  call: any;
}
enum GMPStatus {
  CALL = "call",
  APPROVED = "approved",
  EXECUTED = "executed",
  ERROR = "error",
  GAS_UNPAID = "gas_unpaid",
}
```

### Trigger manual relay of transaction through the Axelar network

The following method, once invoked, will:

1. Query the current status of the transaction to be in one of the states above.
2. Recover from source to destination if needed.

```ts
const txHash =
  "0xfb6fb85f11496ef58b088116cb611497e87e9c72ff0c9333aa21491e4cdd397a";
const src = "Ethereum";
const dest = "Avalanche";
const debug = true;
const recover = await api.manualRelayToDestChain({ txHash, src, dest, debug });
```

Possible return values are: - `already executed` - Transaction was already executed and a manual recovery was not necessary. - `triggered relay` - The `manualRelayToDestChain` trigggered a manual relay through our network. - `approved but not executed` - The transaction already reached the destination chain but was not executed to reach the intended destination contract address. - => WHEN IN THIS STATE, THERE ARE TWO OPTIONS TO REMEDIATE (BELOW).

### Execute manually OR increase gas payment

#### 1. Execute manually

When invoking this method, you will manually execute (and pay for) the executable method on your specified contract on the destination chain of your cross-chain transaction.

```ts
// TODO: the txState query can be improved
const testnetCachingServiceAPI: string =
  "https://testnet.api.gmp.axelarscan.io";
const txState = await api.execGet(testnetCachingServiceAPI, {
  method: "searchGMP",
  txHash,
});
await sdk.executeManually(res[0], (data: any) => console.log(data));
```

Possible return values are:

```ts
{
    status: "pending" | "success" | "failed",
    message: "Wait for confirmation" | "Execute successful" | <ERROR>,
    txHash: tx.hash,
}
```

#### 2. Increase Gas Payment

There're two different functions to increase gas payment depending on type of the token.

##### 2.1 Native Gas Payment

Invoking this method will execute the `addNativeGas` method on the gas receiver contract on the source chain of your cross-chain transaction to increase the amount of the gas payment, in the form of **native token**. The amount to be added is automatically calculated based on many factors e.g. token price of the destination chain, token price of the source chain, current gas price at the destination chain, etc. However, it can be overrided by specifying amount in the `options`.

```ts
import {
  AxelarGMPRecoveryAPI,
  Environment,
  AddGasOptions,
} from "@axelar-network/axelarjs-sdk";

// Optional
const options: AddGasOptions = {
  amount: "10000000", // Amount of gas to be added. If not specified, the sdk will calculate the amount automatically.
  refundAddress: "", // If not specified, the default value is the sender address.
  estimatedGasUsed: 700000, // An amount of gas to execute `executeWithToken` or `execute` function of the custom destination contract. If not specified, the default value is 700000.
  evmWalletDetails: { useWindowEthereum: true, privateKey: "0x" }, // A wallet to send an `addNativeGas` transaction. If not specified, the default value is { useWindowEthereum: true}.
};

const txHash: string = "0x...";
const { success, transaction, error } = await api.addNativeGas(
  EvmChain.AVALANCHE,
  txHash,
  options
);

if (success) {
  console.log("Added native gas tx:", transaction?.transactionHash);
} else {
  console.log("Cannot add native gas", error);
}
```

##### 2.2 ERC-20 Gas Payment

This is similar to native gas payment except using **ERC-20 token** for gas payment. However, the supported ERC-20 tokens are limited. See the list of supported tokens here: [[Mainnet](/resources/mainnet) | [Testnet](/resources/testnet) | [Testnet-2](/resources/testnet-2)]

```ts
import {
  AxelarGMPRecoveryAPI,
  Environment,
  AddGasOptions,
  EvmChain,
  GAS_RECEIVER,
} from "@axelar-network/axelarjs-sdk";
import { ethers } from "ethers";

// Optional
const options: AddGasOptions = {
  amount: "10000000", // The amount of gas to be added. If not specified, the sdk will calculate the amount to be paid.
  refundAddress: "", // The address to get refunded gas. If not specified, the default value is the tx sender address.
  estimatedGasUsed: 700000, // An amount of gas to execute `executeWithToken` or `execute` function of the custom destination contract. If not specified, the default value is 700000.
  evmWalletDetails: { useWindowEthereum: true, privateKey: "0x" }, // A wallet to send an `addNativeGas` transaction. If not specified, the default value is { useWindowEthereum: true}.
};

const environment = Environment.TESTNET; // Can be `Environment.TESTNET` or `Environment.MAINNET`
const api = new AxelarGMPRecoveryAPI({ environment });

// Approve gas token to the gas receiver contract
const gasToken = "0xGasTokenAddress";
const erc20 = new ethers.Contract(gasToken, erc20Abi, gasPayer);
await erc20
  .approve(GAS_RECEIVER[environment][EvmChain.AVALANCHE], amount)
  .then((tx) => tx.wait());

// Send `addGas` transaction
const { success, transaction, error } = await api.addGas(
  EvmChain.AVALANCHE,
  "0xSourceTxHash",
  gasToken,
  options
);

if (success) {
  console.log("Added gas tx:", transaction?.transactionHash);
} else {
  console.log("Cannot add gas", error);
}
```
