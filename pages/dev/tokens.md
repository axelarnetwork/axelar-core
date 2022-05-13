# Transfer tokens cross-chain

There are two ways to transfer tokens cross-chain with Axelar:

1. Call `sendToken` on an Axelar gateway EVM contract.
2. Get a deposit address using the AxelarJS SDK.

Use `sendToken` if:

- Your app transfers EVM-to-X where X is one of EVM, Cosmos.
- Your app uses smart contracts.

Use a deposit address if:

- You need functionality not offered by `sendToken`. Example: Cosmos-to-X.
- You want to allow token transfers from wallets that don't know anything about Axelar. Example: Withdrawal from a centralized exchange.

## Call `sendToken`

### Overview

1. Locate the Axelar Gateway contract on the source chain
2. Execute approve on the source chain (ERC-20)
3. Execute sendToken on the Gateway

### 1. Locate the Axelar Gateway contract on the source chain

Axelar Gateways are application-layer smart contracts established on source and destination chains. They send and receive payloads, and monitor state. Find a list of gateway addresses for the chains we support in Resources [[Mainnet](../resources/mainnet) | [Testnet](../resources/testnet) | [Testnet-2](../resources/testnet-2)].

An Axelar Gateway implements the `IAxelarGateway` interface, which has a public method called `sendToken`:

```solidity
function sendToken(
    string memory destinationChain,
    string memory destinationAddress,
    string memory symbol,
    uint256 amount
) external;
```

### 2. Execute approve on the source chain (ERC-20)

Transferring tokens through a Gateway is similar to an ERC-20 token transfer. You first need to approve the Gateway to transfer a specific token in a specific amount. This approval is done via the `approve` method of the ERC-20 interface:

```solidity
function approve(address spender, uint256 amount) external returns (bool);
```

Here `spender` is the Gateway address on the source chain.

Find a list of assets, their names and their addresses in Resources [[Mainnet](../resources/mainnet) | [Testnet](../resources/testnet) | [Testnet-2](../resources/testnet-2)].

### 3. Execute sendToken on the Gateway

Call `sendToken` on the gateway contract of the source chain. Example:

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

A _deposit address_ is a special address created and monitored by Axelar relayer services. It is similar to how centralized exchanges generate a monitored one-time deposit address that facilitate your token transfers.

Deposit address workflow:

1. Generate a deposit address.
2. User sends tokens to the deposit address. Examples: withdrawal from a centralized exchange, transaction from your favorite wallet software.
3. Axelar relayers observe the transaction on the source chain and complete it on the destination chain.

### Install the AxelarJS SDK

We'll use the AxelarJS SDK, which is a `npm` dependency that empowers developers to make requests into the Axelar network from a frontend. The Axelar SDK provides a wrapper for API calls that you can use to generate a deposit address. (Alternately, see [Send UST to an EVM chain](../learn/cli/ust-to-evm) for an example of how to generate a deposit address using the CLI instead of the Axelar SDK.)

Install the Axelar SDK:

```bash
npm i @axelar-network/axelarjs-sdk@alpha
```

(We recommend installing the alpha version as the stable version is now legacy. The alpha version is scalable and uses websockets to wait for the deposit address generation.)

### Generate a deposit address using the SDK

Call `getDepositAddress`:

```tsx
async getDepositAddress(
  fromChain: string, // source chain
  toChain: string, // destination chain
  destinationAddress: string, // destination address to transfer the token to
  asset: string, // common key of the asset
  options?: {
    _traceId: string;
  }
): Promise<string> {}
```

Example: Cosmos-to-EVM (Terra to Avalanche):

```tsx
const sdk = new AxelarAssetTransfer({
  environment: "testnet",
  auth: "local",
});
const depositAddress = await sdk.getDepositAddress(
  "terra", // source chain
  "avalanche", // destination chain
  "0xF16DfB26e1FEc993E085092563ECFAEaDa7eD7fD", // destination address
  "uusd" // asset to transfer
);
```

Example: EVM-to-Cosmos (Avalanche to Terra)

```tsx
const sdk = new AxelarAssetTransfer({
  environment: "testnet",
  auth: "local",
});
const depositAddress = await sdk.getDepositAddress(
  "avalanche", // source chain
  "terra", // destination chain
  "terra1qem4njhac8azalrav7shvp06myhqldpmkk3p0t", // destination address
  "uusd" // asset to transfer
);
```

Note: The destination address format is validated based on the destination chain. Make sure that the destination address is a valid address on the destination chain. For instance Terra addresses start with “terra,” Osmosis with “osmo,” etc.

Once the deposit address has been generated the user can make a token transfer (on blockchain) to the deposit address. The transfer will be picked up by the Axelar network and relayed to the destination chain.
