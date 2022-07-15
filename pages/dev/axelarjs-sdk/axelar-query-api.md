# Axelar Query API

This module is a nice-to-have of common queries made into the Axelar network and its services that are abstracted into easy-to-use Javascript one-liners.

### Install the AxelarJS SDK module (AxelarQueryAPI)

Install the AxelarJS SDK:

```bash
npm i @axelar-network/axelarjs-sdk@alpha
```

Instantiate the `AxelarQueryAPI` module:

```ts
const sdk = new AxelarQueryAPI({
  environment: "testnet",
});
```

### Possible queries

#### estimateGasFee

Useful query for GMP transactions, when invoking `callContract` or `callContractWithToken` to get an estimate of the appropriate gas payment to be made to the gas receiver on the source chain.

```ts
// (Optional) An estimated gas amount required to execute `executeWithToken` function. The default value is 700000 which sufficients for most transaction.
const estimateGasUsed = 400000;

// Returns avax amount to pay gas
const gasFee = await sdk.estimateGasFee(
  EvmChain.AVALANCHE,
  EvmChain.FANTOM,
  GasToken.AVAX,
  estimateGasUsed
);
```

#### getTransferFee

Given a source chain, destination chain, and amount of an asset, retrieves the base fee that the network would assess for that transaction

```ts
/**
 * Gets the transfer fee for a given transaction
 * @param sourceChainName
 * @param destinationChainName
 * @param assetDenom
 * @param amountInDenom
 * @returns
 */
public async getTransferFee(
  sourceChainName: string,
  destinationChainName: string,
  assetDenom: string,
  amountInDenom: number
): Promise<TransferFeeResponse>
```

#### getFeeForChainAndAsset

Given a chain and asset, retrieves the base fee that the network would assess

```ts
/**
 * Gets the fee for a chain and asset
 * @param chainName
 * @param assetDenom
 * @returns
 */
public async getFeeForChainAndAsset(
  chainName: string,
  assetDenom: string
): Promise<FeeInfoResponse>
```

#### getDenomFromSymbol

Get the denom for an asset given its symbol on a chain

```ts
/**
 * @param symbol
 * @param chainName
 * @returns
 */
public getDenomFromSymbol(symbol: string, chainName: string)
```

#### getSymbolFromDenom

Get the symbol for an asset on a given chain given its denom

```ts
/**
 * @param denom
 * @param chainName
 * @returns
 */
public getSymbolFromDenom(denom: string, chainName: string)
```
