# Call a contract on chain B from chain A

import Callout from 'nextra-theme-docs/callout'

To call a contract on chain B from chain A, the user needs to call `callContract` on the gateway of chain A, specifying:

- The destination chain, which must be an EVM chain from [Chain names](./chain-names).
- The destination contract address, which must implement the `IAxelarExecutable` interface defined in [IAxelarExecutable.sol](https://github.com/axelarnetwork/axelar-cgp-solidity/blob/main/contracts/interfaces/IAxelarExecutable.sol).
- The payload `bytes` to pass to the destination contract.

```solidity
function callContract(
    string memory destinationChain,
    string memory contractAddress,
    bytes memory payload
) external;
```

`IAxelarExecutable` has an `_execute` function that will be triggered by the Axelar network after the `callContract` function has been executed. You can write any custom logic there.

```solidity
function _execute(
    string memory sourceChain,
    string memory sourceAddress,
    bytes calldata payload
) internal virtual {}
```

<Callout emoji="ℹ️">
  Ensure the payload is encoded `bytes`!
</Callout>

The `payload` passed to `callContract` (and ultimately to `_execute` and `_executeWithToken`) has type `bytes`. Use the ABI encoder/decoder convert your data to `bytes`.

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
