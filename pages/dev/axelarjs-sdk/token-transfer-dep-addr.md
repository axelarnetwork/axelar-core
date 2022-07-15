
# Get a deposit address

A _deposit address_ is a special address created and monitored by Axelar relayer services on behalf of the requester. It is similar to how centralized exchanges generate a monitored, one-time deposit address that facilitates your token transfers.

### Deposit address workflow:

1. Generate a deposit address on a specific source chain.
2. User sends tokens to the deposit address on the source chain. Examples: withdrawal from a centralized exchange, transaction from your favorite wallet software.
3. Axelar relayers observe the deposit transaction on the source chain and complete it on the destination chain.
4. Watch your tokens arrive on the destination chain.

### 1. Install the AxelarJS SDK module (AxelarAssetTransfer)

We'll use the AxelarJS SDK, which is an `npm` dependency that empowers developers to make requests into the Axelar network from a front end. The Axelar SDK provides a wrapper for API calls that you can use to generate a deposit address. (Alternately, you can generate a deposit address using the CLI instead of the Axelar SDK. [See examples, here](../../learn/cli).) 

1. Install the AxelarJS SDK:

```bash
npm i @axelar-network/axelarjs-sdk
```

2. Instantiate the `AxelarAssetTransfer` module:

```bash
const sdk = new AxelarAssetTransfer({
  environment: "testnet",
  auth: "local",
});
```

### 2. Generate a deposit address using the SDK

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

Example: Cosmos to EVM (Osmosis to Avalanche):

```tsx
const sdk = new AxelarAssetTransfer({
  environment: "testnet"
});
const depositAddress = await sdk.getDepositAddress(
  "osmosis", // source chain
  "avalanche", // destination chain
  "0xF16DfB26e1FEc993E085092563ECFAEaDa7eD7fD", // destination address
  "uausdc" // asset to transfer in atomic denom units
);
```

Example: EVM to Cosmos (Avalanche to Osmosis)

```tsx
const sdk = new AxelarAssetTransfer({
  environment: "testnet",
  auth: "local",
});
const depositAddress = await sdk.getDepositAddress(
  "avalanche", // source chain
  "osmosis", // destination chain
  "osmo1x3z2vepjd7fhe30epncxjrk0lehq7xdqe8ltsn", // destination address
  "uausdc" // asset to transfer in atomic denom units
);
```

Note: The destination address format is validated based on the destination chain. Make sure the destination address is a valid address on the destination chain. For instance, Osmosis with “osmo,” etc.

Once the deposit address has been generated, the user can make a token transfer (on blockchain) to the deposit address. The transfer will be picked up by the Axelar network and relayed to the destination chain.
