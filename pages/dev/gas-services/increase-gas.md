# Increase gas

On occasion, the prepaid gas to the Gas Receiver contract could be insufficient, such as when the destination chain is congested with many transfers or due to other conditions. Therefore, Axelar provides an option to resubmit a new amount of gas, as well as an option to refund the paid gas. 

The process can be done through the Axelarscan UI, the Axelar SDK or via direct invocation of the Gas Receiver contract.

When needed, our smart contract can receive gas to top up an existing General Message Passing (GMP) transaction in the following ways:

### `addNativeGas`

Pay additional gas to a transaction that needs it (referenced by its txHash) in native tokens on its EVM source chain.

- In Solidity:

```solidity
function addNativeGas(
    bytes32 txHash,
    uint256 logIndex,
    address refundAddress
) external payable override;
```

- In JavaScript or TypeScript, the SDK abstracts a method that can be invoked directly in a web application.
  See SDK docs for [Increase Native Gas Payment](/dev/axelarjs-sdk/tx-status-query-recovery#21-native-gas-payment).

### `addGas`

Pay additional gas to a transaction that needs it (referenced by its txHash) in any of Axelar's supported tokens on its EVM source chain.

- In Solidity:

```solidity
function addGas(
    bytes32 txHash,
    uint256 logIndex,
    address gasToken,
    uint256 gasFeeAmount,
    address refundAddress
) external override;
```

- In JavaScript or TypeScript: [Increase ERC-20 Gas Payment](/dev/axelarjs-sdk/tx-status-query-recovery#22-erc-20-gas-payment).

\*\*\* **Can only be paid in tokens that Axelar supports. See the list of supported assets for the chains we support: [Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet).** \*\*\* 
