# `Pay Gas`

For both `callContract` and `callContractWithToken`, the Gas Receiver can be paid in the native token of the source chain or any token supported by Axelar network. 

## For `callContract`
### `payGasForContractcall`:
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

*** Gas Receiver can only be paid in tokens Axelar supports. See the list of supported assets in Resources [[Mainnet](/resources/mainnet) | [Testnet](/resources/testnet) | [Testnet-2](/resources/testnet-2)].


### `payGasForContractCallWithToken`

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

## For `callContractWithToken`:
### `payGasForContractCallWithToken`
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

*** Gas Receiver can only be paid in tokens Axelar supports. See the list of supported assets in Resources [[Mainnet](/resources/mainnet) | [Testnet](/resources/testnet) | [Testnet-2](/resources/testnet-2)].

### `payNativeGasForContractCallWithToken`
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

The function names are prety self-explanatory. The following is true for the arguments:

- For all functions
  - `sender` needs to match the address that calls `callContract` or `callContractWithToken` on the `AxelarGateway`. If the `AxelarGasReceiver` is called by the same contract that will call the Gateway then simply specify `address(this)` as `sender`.
- For `payGasForContractCall` and `payNativeGasForContractCall`:
  - `destinationChain`
  - `destinationAddress`
  - `payload`
    need to match the arguments of a `contractCall` on the `AxelarGateway`.
- For `payGasForContractCallWtihToken` and `payNativeGasForContractCallWithToken`:
  - `destinationChain`
  - `destinationAddress`
  - `payload`
  - `symbol`
  - `amount`
    need to match the arguments of a `contractCallWithToken` on the `AxelarGateway`.
- For `payGasForContractCall` and `payGasForContractCallWtihToken`:
  - `gasToken` is the address of the token that gas will be paid in. Ensure this token is supported by the network, using the Axelar API.
  - `gasFeeAmount` is the amount of `gasToken` to transfer from the sender. The sender needs to have approved the `AxelarGasReceiver` with the appropriate amount to `gasToken` first.
- For `payNativeGasForContractCall` and `payNativeGasForContractCallWithToken`, the amount of funds received is specified by `msg.value`.
- For all functions,
  - `refundAddress` is the address that will eventually be able to receive excess amount paid for gas.
