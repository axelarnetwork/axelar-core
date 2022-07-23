# Transfer tokens cross-chain

There are two ways to transfer tokens cross-chain with Axelar:

- Call `sendToken` on an Axelar gateway EVM contract.
- Get a deposit address using the [AxelarJS SDK](../axelarjs-sdk/token-transfer-dep-addr).

## Call `sendToken`

Use `sendToken` if:

- Your dApp transfers EVM-to-X where X is one of EVM, Cosmos.
- Your dApp uses smart contracts.

### Overview

1. Locate the Axelar Gateway contract on the source chain.
2. Execute approve on the source chain (ERC-20).
3. Execute sendToken on the Gateway.

#### 1. Locate the Axelar Gateway contract on the source chain

Axelar Gateways are application-layer smart contracts established on source and destination chains. They send and receive payloads, and monitor state. Find a list of Gateway addresses for the chains we support: [Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet).

An Axelar Gateway implements the `IAxelarGateway` interface, which has a public method called `sendToken`:

```solidity
function sendToken(
    string memory destinationChain,
    string memory destinationAddress,
    string memory symbol,
    uint256 amount
) external;
```

#### 2. Execute approve on the source chain (ERC-20)

Transferring tokens through a Gateway is similar to an ERC-20 token transfer. You first need to approve the Gateway to transfer a specific token in a specific amount. This approval is done via the `approve` method of the ERC-20 interface:

```solidity
function approve(address spender, uint256 amount) external returns (bool);
```

Here, `spender` is the Gateway address on the source chain.

Find a list of assets, their names and their addresses: [Mainnet](../build/contract-addresses/mainnet) | [Testnet](../build/contract-addresses/testnet).

#### 3. Execute sendToken on the Gateway

Call `sendToken` on the Gateway contract of the source chain. Example:

```solidity
sendToken(
    "avalanche", // destination chain name
    "0xF16DfB26e1FEc993E085092563ECFAEaDa7eD7fD", // some destination wallet address (should be your own)
    "USDC", // asset symbol
    100000000 // amount (in atomic units)
)
```

Watch for the tokens to appear at the destination address on the destination chain.

## Get a deposit address
Use a deposit address if:

- You need functionality not offered by `sendToken`. Example: Cosmos-to-X.
- You want to allow token transfers from wallets that don't know anything about Axelar. Example: Withdrawal from a centralized exchange.


Refer to [AxelarJS SDK](../axelarjs-sdk/token-transfer-dep-addr).
