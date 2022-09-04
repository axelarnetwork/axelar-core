
# Call a contract on chain B from chain A and attach some tokens

import Callout from 'nextra-theme-docs/callout'

To call chain B from chain A and send some tokens along the way, the user needs to call `callContractWithToken` on the gateway of chain A, specifying:

- The destination chain, which must be an EVM chain from [Chain names](./chain-names).
- The destination contract address, which must implement the `IAxelarExecutable` interface defined in [IAxelarExecutable.sol](https://github.com/axelarnetwork/axelar-cgp-solidity/blob/main/contracts/interfaces/IAxelarExecutable.sol).
- The payload `bytes` to pass to the destination contract.
- The symbol of the token to transfer, which must be a supported asset ([Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet)).
- The amount of the token to transfer.

As per the snippet below:

```solidity
function callContractWithToken(
    string memory destinationChain,
    string memory contractAddress,
    bytes memory payload,
    string memory symbol,
    uint256 amount
) external;
```

`IAxelarExecutable` has an `_executeWithToken` function that will be triggered by the Axelar network after the `callContractWithToken` function has been executed. You can write any custom logic there.

The destination contract will be authorized to transfer the ERC-20 identified by the `tokenSymbol`.

```solidity
function _executeWithToken(
    string memory sourceChain,
    string memory sourceAddress,
    bytes calldata payload,
    string memory tokenSymbol,
    uint256 amount
) internal virtual {}
```

### Example

Suppose our destination contract wants to forward the token it received to a recipient provided in the payload. It could be done this way.

```solidity
function _executeWithToken(
    string memory sourceChain,
    string memory sourceAddress,
    bytes calldata payload,
    string memory tokenSymbol,
    uint256 amount
) internal virtual {
    // decode recipient
    address recipient = abi.decode(payload, (address));
    // get ERC-20 address from gateway
    address tokenAddress = gateway.tokenAddresses(tokenSymbol);

    // transfer received tokens to the recipient
    IERC20(tokenAddress).transfer(recipient, amount);
}
```

<Callout emoji="ℹ️">
  Ensure the payload is encoded `bytes`!
</Callout>

The `payload` passed to `callContract` (and ultimately to the `_execute` and `_executeWithToken`) has type `bytes`. Use the ABI encoder/decoder convert your data to `bytes`.

Example of payload encoding in JavaScript (using ethers.js):

```jsx
const { ethers } = require("ethers");

// encoding a string
const payload = ethers.utils.defaultAbiCoder.encode(
  ["string"],
  ["Hello from contract A"]
);
```

Example of payload decoding in Solidity:

```solidity
function _execute(
    string memory sourceChain,
    string memory sourceAddress,
    bytes calldata payload
) internal override {
    // decoding a string
    string memory _message = abi.decode(payload, (string));
}
```
