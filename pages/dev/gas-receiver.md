# Axelar Gas Receiver

General Message Passing is a two-step process:

1. **_Approval._** Axelar validators approve a message via vote on the message payload hash.
2. **_Execution._** Anyone can execute an approved message by posting the payload hash preimage to the destination chain and paying relayer gas fees.

The Axelar network provides approval but not execution. Axelar also provides an optional relayer service called _gas receiver_ that provides execution of approved messages. Anyone can use the gas receiver service by pre-paying the relayer gas fee on the source chain. Axelar relayer services observe use of the gas receiver for a given message and automatically execute the General Message Passing call.

Relayer gas fees are needed only to pay gas costs across chains.

## Introduction

An application that wants Axelar to automatically execute contract calls on the destination chain needs to do four things:

- Estimate the `gasLimit` that the contract call will require on the destionation chain.
- Query our API to get the `sourceTokenPrice` of the desired token that gas will be paid in, as well as `destinationGasPrice` and `destinationTokenPrice`, the price of the native token, on the destination chain.
- Calcualate the amount of token to be paid as `gasLimit * destinationGasPrice * destinationTokenPrice / sourceTokenPrice`.
- Pay our `AxelarGasReceiver` smart contract on the source chain that amount. This can be done by the application's smart contracts so no additional transactions are required (except maybe approval in case gas is paid in non-native ERC-20 tokens).

Our service does the following:

- Monitors `AxelarGasReceiver` for receipts of payment, and gets the amount paid as `amountPaid`.
- Matches those to contract calls.
- Queries our API to get the `sourceTokenPrice` of the token that gas was paid in, as well as `destinationGasPrice` and `destinationTokenPrice`, the price of the native token, on the destination chain.
- Calcualate the `gasLimit` as `amountPaid * sourceTokenPrice / (destinationGasPrice * gestinationTokenPrice)`.
- Executes the specified contract call specifying the `gasLimit` specified above.

We plan to add an option to get refunds in case excessive amounts are paid as gas, but this is not yet implemented.

## `AxelarGasReceiver`

Our smart contract can receive gas in the following ways:

```solidity
function payGasForContractCall(
    address sender,
    string calldata destinationChain,
    string calldata destinationAddress,
    bytes calldata payload,
    address gasToken,
    uint256 gasFeeAmount,
    address refundAddress
) external;
```

```solidity
// This is called on the source chain before calling the gateway to execute a remote contract.
function payGasForContractCallWithToken(
    address sender,
    string calldata destinationChain,
    string calldata destinationAddress,
    bytes calldata payload,
    string calldata symbol,
    uint256 amount,
    address gasToken,
    uint256 gasFeeAmount,
    address refundAddress
) external;
```

```solidity
// This is called on the source chain before calling the gateway to execute a remote contract.
function payNativeGasForContractCall(
    address sender,
    string calldata destinationChain,
    string calldata destinationAddress,
    bytes calldata payload,
    address refundAddress
) external payable;
```

```solidity
// This is called on the source chain before calling the gateway to execute a remote contract.
function payNativeGasForContractCallWithToken(
    address sender,
    string calldata destinationChain,
    string calldata destinationAddress,
    bytes calldata payload,
    string calldata symbol,
    uint256 amount,
    address refundAddress
) external payable;
```

The function names are prety self explanatory. The following is true for the arguments:

- For all functions
  - `sender` needs to match the address that calls `callContract` or `callContractWithToken` on the `AxelarGateway`. If the `AxelarGasReceiver` is called by the same contract that will call the gateway then simply specify `address(this)` as `sender`.
- For `payGasForContractCall` and `payNativeGasForContractCall`
  - `destinationChain`
  - `destinationAddress`
  - `payload`
    need to match the arguments of a `contractCall` on the `AxelarGateway`
- For `payGasForContractCallWtihToken` and `payNativeGasForContractCallWithToken`
  - `destinationChain`
  - `destinationAddress`
  - `payload`
  - `symbol`
  - `amount`
    need to match the arguments of a `contractCallWithToken` on the `AxelarGateway`
- For `payGasForContractCall` and `payGasForContractCallWtihToken`
  - `gasToken` is the address of the token that gas will be paid in. Ensure this token is supported with our API.
  - `gasFeeAmount` is the amount of `gasToken` to transfer from the sender. The sender needs to have approved the `AxelarGasReceiver` with the appropriate amount to `gasToken` first.
- For `payNativeGasForContractCall` and `payNativeGasForContractCallWithToken` the amount of funds received is specified by `msg.value`.
- For all functions
  - `refundAddress` is the address that will be able to receive excess amount paid for gas.

## API

To get the relative gas cost in any token on any of the supported chains you can use the following script. This is subject to change.

```ts
const {
  constants: { AddressZero },
} = require("ethers");
const axios = require("axios");

async function getGasPrice(
  sourceChain,
  destinationChain,
  tokenAddress,
  tokenSymbol
) {
  const api_url = "https://devnet.api.gmp.axelarscan.io";

  const requester = axios.create({ baseURL: api_url });
  const params = {
    method: "getGasPrice",
    destinationChain: destinationChain,
    sourceChain: sourceChain,
  };

  // set gas token address to params
  if (tokenAddress != AddressZero) {
    params.sourceTokenAddress = tokenAddress;
  } else {
    params.sourceTokenSymbol = tokenSymbol;
  }
  // send request
  const response = await requester.get("/", { params }).catch((error) => {
    return { data: { error } };
  });
  const result = response.data.result;
  const dest = result.destination_native_token;
  const destPrice = 1e18 * dest.gas_price * dest.token_price.usd;
  return destPrice / result.source_token.token_price.usd;
}
```
