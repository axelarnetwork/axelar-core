# Axelar Gas Services

For any General Message Passing (GMP) transaction, the Axelar network routes transctions to their destination chains. The final step of the pipeline (execution) to the specified destination contract address on the destination chain is invoked in one of two ways.

1. Manually paid by the user/application on the destination chain.
2. Executed automatically by Axelar if the user/application prepays gas to the Gas Receiver contract on the source chain.

Gas Receiver is a smart contract deployed on every EVM that is provided by Axelar. Our Gas Services provide users/applications the ability to prepay the full estimated cost for any GMP transaction (from source to destination chains) with the convenience of a single payment on the source chain, relying on Axelar's relay services to manage the full pipeline. Once gas is paid to our gas services for a GMP transaction, Axelar's relayer services pick up the payment and automatically execute the final General Message Passing call.

Developers can use the Gas Services by prepaying upfront the relayer gas fee on the source chain, thereby covering the cost of gas to execute the final transaction on the destination chain.

## Overview of Gas Services

### Pay gas

An application that wants Axelar to automatically execute contract calls on the destination chain needs to do four things:

1. Estimate the `gasLimit` that the contract call will require on your executable contract on the destination chain.

2. Call the `estimateGasFee` method to get the `sourceGasFee` in the desired gas-payment token on the destination chain. See [[Sample code](/dev/axelarjs-sdk/axelar-query-api#estimategasfee)] for reference.

Prerequisite: Axelar SDK must be installed. Refer to [[AxelarJS SDK](/dev/axelarjs-sdk/intro)].

3. Calculate the amount of token to be paid.
   `gasLimit` \* `sourceGasPrice`.

4. Pay the AxelarGasService smart contract on the source chain in the amount calculated in step 3.

For example, we want to interact with the contract below

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

So, in the front-end side, we need to pass `sourceGasFee` to `msg.value` like below

```ts
await contract.sendToMany("moonbeam", "0x...", ["0x.."], "USDC", 1, {
  value: sourceGasFee, // This is the value we get from Step 3.
});
```

After sending a transaction out, our service will do the following:

- Monitors `AxelarGasReceiver` for receipt of payment, and gets the amount paid as `amountPaid`.
- Matches those to contract calls.
- Executes the specified contract call, specifying the `gasLimit` defined above.

See [[Pay Gas](pay-gas)] for more.

### Increase Gas

The prepaid gas to the Gas Service contract (in GMP step 2) could be insufficient when the destination chain is congested (with many transfers or other conditions). Therefore, Axelar provides an option to resubmit a new amount of gas, as well as an option to refund the paid gas. The process can be done through the Axelarscan UI, the Axelar SDK, or via direct invocation of the Gas Receiver contract.

See [[Increase Gas](increase-gas)] for more.

### Refund Gas

We plan to add an option to get refunds in case excessive amounts are paid as gas, but this is not yet implemented.
