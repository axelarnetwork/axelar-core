# AxelarJS SDK

The AxelarJS SDK is an npm library that includes a collection of APIs and query tools written in Javascript. 

The package can be installed into your dApp as a project dependency with:
```bash
npm i @axelar-network/axelarjs-sdk@alpha
```


Current modules:

- `AxelarAssetTransfer`.
    - Used for cross-chain token transfer via deposit address generation.
    - [[Token Transfer via Deposit Address](token-transfer-dep-addr)].

- `AxelarGMPRecoveryAPI`.
    - API library to track and recover (if needed) GMP transactions (both `callContract` and `callContractWithToken`).
    - Transactions are indexed by the transaction hash initiated on the source chain when invoking either `callContract` or `callContractWithToken`.
    - [[GMP transaction status and recovery](tx-status-query-recovery)].

- `AxelarQueryAPI`.
    - Collection of helpful predefined queries into the network, e.g. for transaction fees for token transfers, cross-chain gas prices for GMP transactions, denom conversions, etc.
    - [[Axelar Query API](axelar-query-api)].