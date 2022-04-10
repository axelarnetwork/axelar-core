# General message passing

import Callout from 'nextra-theme-docs/callout'

<Callout type="warning" emoji="⚠️">
  Under construction
</Callout>

Consider unrelated EVM chains A and B (the source and destination chain). It is assumed that gateway contracts run on both chains A and B. Three different scenarios are as follows:

- Sending token from chain A to chain B
- Calling a contract on chain B from chain A
- Calling a contract on chain B from chain A with some token attached

### Sending token from chain A to chain B

In order to send a token directly from the Gateway on chain A user or smart contract would need to call a following method:

```solidity
function sendToken(
    string memory destinationChain,
    string memory destinationAddress,
    string memory symbol,
    uint256 amount
)
```

This would result in instant burning of the token on chain A and emitting an event `TokenSent` from the Gateway on chain A. Axelar network will listen to such events and then specified token will be minted on the destination chain B. Some fee might be subtracted from the amount in the process

### Calling a contact on chain B from chain A

To calling a contact on chain B from directly from the Gateway on chain A user or smart contract would need to call a following method:

```solidity
function callContract(
    string memory destinationChain,
    string memory destinationContractAddress,
    bytes memory payload
)
```

An event called `ContractCall` will be emitted and Axelar network will listen to it and then send a command called `approveContractCall` to the gateway on the chain B. In order to utilize that approval a 3rd party contract on chain B would need to implement the `execute` method which is calling the gateway method `validateContractCall` to validate the approval

### Calling a contract on chain B from chain A with some token attached

To calling a contact on chain B from directly from the Gateway on chain A with some token attached the following method needs to be called:

```solidity
function callContractWithToken(
    string memory destinationChain,
    string memory destinationContractAddress,
    bytes memory payload,
    string memory symbol,
    uint256 amount
)
```

An event called `ContractCallWithToken` will be emitted and Axelar network will listen to it and then send a command called `approveContractCallWithMint` to the gateway on the chain B. In order to utilize that approval a 3rd party contract on chain B would need to implement the `executeWithToken` method which is calling the gateway method `validateContractCallAndMint` to validate the approval and have the related token minted to that 3rd party contract. From there that contract can execute arbitrary logic, talk to any existing contracts and use the minted token in any ways.
