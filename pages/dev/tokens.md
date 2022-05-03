# Transfer tokens cross-chain

There are two ways to transfer tokens cross-chain with Axelar:

1. Call `sendToken` on an Axelar gateway EVM contract.
2. Get a deposit address using the AxelarJS SDK.

Use `sendToken` if:

- Your app transfers EVM-to-X where X is one of EVM, Cosmos.
- Your app uses smart contracts.

Use a deposit address if:

- You need functionality not offered by `sendToken`. Example: Cosmos-to-X.
- You want to accept withdrawals from a centralized exchange.

## Call `sendToken`

### Outline

1. Locate the Axelar Gateway contract on the source chain
2. Execute approve on the source chain (ERC-20)
3. Execute sendToken on the Gateway

### 1. Locate the Axelar Gateway contract on the source chain

Axelar Gateways are application-layer smart contracts established on source and destination chains. They send and receive payloads, and monitor state. Find a list of gateway addresses for the chains we support in [Resources](../resources).

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

Find a list of assets, their names and their addresses in [Resources](../resources).

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

The Axelar SDK is a wrapper around API calls that you can use to generate a deposit address. A _deposit address_ is a special address created and monitored by Axelar. It is similar to how crypto exchanges generate a monitored one-time deposit address that facilitate your crypto transfers.

The process using the Axelar SDK is as follows:

1. User specifies the source chain, destination chain and the asset to be transferred.
2. Axelar generates a deposit address.
3. User sends tokens to the deposit address. Examples: withdrawal from a centralized exchange, transaction from your favorite wallet software.
4. Axelar picks up the transfer and forwards it to the destination chain.

### 1. Install Axelar SDK

The Axelar SDK is a JavaScript npm package. Install the package:

```bash
npm i @axelar-network/axelarjs-sdk@alpha
```

(We recommend installing the alpha version as the stable version is now legacy. The alpha version is scalable and uses websockets to wait for the deposit address generation.)

### 2. Generate a deposit address with the SDK

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
    environment: 'testnet',
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
    environment: 'testnet',
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
