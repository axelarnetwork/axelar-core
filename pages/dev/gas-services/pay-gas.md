# Pay gas

* [Overview](/dev/gas-services/pay-gas#overview)
* [Example](/dev/gas-services/pay-gas#example)
* [Alternative gas payment methods for callContract](/dev/gas-services/pay-gas#alternatives-for-paying-gas-for-callcontract)
* [Alternative gas payment methods for callContractWithToken](/dev/gas-services/pay-gas#alternatives-for-paying-gas-for-callcontractwithtoken)
* [Two-way Call](/dev/gas-services/pay-gas#two-way-call)
## Overview 
Axelar gas services provide methods to pay the relayer gas fee for both `callContract` and `callContractWithToken`. The fee can be paid in the native token of the source chain or any token supported by Axelar network. An application that wants Axelar to automatically execute contract calls on the destination chain needs to do three things:

1. Estimate the `gasLimit` that the contract call will require on your executable contract on the destination chain.

2. Call the `estimateGasFee` method to get the `sourceGasFee` in the desired gas-payment token on the destination chain. See this [code snippet](/dev/axelarjs-sdk/axelar-query-api#estimategasfee) for reference. (As a prerequisite, Axelar SDK must be installed. Refer to [AxelarJS SDK](/dev/axelarjs-sdk/intro).)

3. Pay the AxelarGasService smart contract on the source chain in the amount calculated in step 2.

## Example
For example, assume the following smart contract is deployed on a source chain:

```solidity
contract SimpleTransferContract {
  ...
  function sendToMany(
      string memory destinationChain,
      string memory destinationContractAddress,
      address[] calldata destinationAddresses,
      string memory symbol,
      uint256 amount
  ) external payable {
      address tokenAddress = gateway.tokenAddresses(symbol);
      IERC20(tokenAddress).transferFrom(msg.sender, address(this), amount);
      IERC20(tokenAddress).approve(address(gateway), amount);
      bytes memory payload = abi.encode(destinationAddresses);

      if(msg.value > 0) {
          // The line below is where we pay the gas fee
          gasReceiver.payNativeGasForContractCallWithToken{value: msg.value}(
              address(this),
              destinationChain,
              destinationContractAddress,
              payload,
              symbol,
              amount,
              msg.sender
          );
      }
      gateway.callContractWithToken(destinationChain, destinationContractAddress, payload, symbol, amount);
  }
}
```

The `msg.value` is the gas amount we pay to the `AxelarGasService` contract.

So, on the front-end side, we need to pass `sourceGasFee` to `msg.value` like below:

```ts
await contract.sendToMany("moonbeam", "0x...", ["0x.."], "USDC", 1, {
  value: sourceGasFee, // This is the value we get from Step 2.
});
```

After sending a transaction out, our Executor Service will do the following:

- Monitor `AxelarGasReceiver` for receipt of payment, and get the amount paid as `amountPaid`.
- Match those to contract calls.
- Execute the specified contract call, specifying the `gasLimit` defined above.

## Alternative gas payment methods for `callContract`
There are two available methods to pay gas for relaying `callContract`. You can choose the one that fits your application design.

### payGasForContractcall
This method receives any tokens for the relayer fee. The paid gas for this method must be in tokens Axelar supports. See the list of supported assets for the chains we support: [Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet).

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

### payNativeGasForContractCall
This method accepts the native tokens of the source chain.

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

## Alternative gas payment methods for `callContractWithToken`
Similar to the available gas payment methods for `callContract`, there are two available methods to pay gas for relaying the `callContractWithToken`.

### payGasForContractCallWithToken
This method receives any tokens for the relayer fee. The paid gas for this method must be in tokens Axelar supports. See the list of supported assets: [Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet).

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

### payNativeGasForContractCallWithToken
This method accepts the native tokens of the source chain.

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

### Methods arguments description
The function names are prety self-explanatory. The following is true for the arguments:

- For all functions:
  - `sender` needs to match the address that calls `callContract` or `callContractWithToken` on the `AxelarGateway`. If the `AxelarGasReceiver` is called by the same contract that will call the Gateway, then simply specify `address(this)` as `sender`.
- For `payGasForContractCall` and `payNativeGasForContractCall`, the following need to match the arguments of a `contractCall` on the `AxelarGateway`:
  - `destinationChain`
  - `destinationAddress`
  - `payload`
- For `payGasForContractCallWtihToken` and `payNativeGasForContractCallWithToken`, the following need to match the arguments of a `contractCallWithToken` on the `AxelarGateway`:
  - `destinationChain`
  - `destinationAddress`
  - `payload`
  - `symbol`
  - `amount`
- For `payGasForContractCall` and `payGasForContractCallWtihToken`:
  - `gasToken` is the address of the token that gas will be paid in. Ensure this token is supported by the network, using the Axelar API.
  - `gasFeeAmount` is the amount of `gasToken` to transfer from the sender. The sender needs to have approved the `AxelarGasReceiver` with the appropriate amount to `gasToken` first.
- For `payNativeGasForContractCall` and `payNativeGasForContractCallWithToken`, the amount of funds received is specified by `msg.value`.
- For all functions, `refundAddress` is the address that will eventually be able to receive excess amount paid for gas.

## Two-way call
The Executor Service supports two-way call, where a message is sent from a source chain, immediately executed at a destination chain, and another message is returned back to the source chain.

The service monitors if there's another contract call immediately executed within the same executed transaction on the destination (to send a message back to the source chain). The service will then automatically use the remaining prepaid gas to relay the second call of the two-way call. 

In case the remaining gas amount is insufficient for the second call, the service refunds it to the payer's address. Users can still pay a new gas amount to relay the second call of the transfer through the [Axelar SDK](/dev/axelarjs-sdk/tx-status-query-recovery#2-increase-gas-payment) or [Axelarscan UI](/dev/monitor-recover/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain).