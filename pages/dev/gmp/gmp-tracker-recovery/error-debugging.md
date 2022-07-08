# Execution Error Messages

import Callout from 'nextra-theme-docs/callout'

Below are some possible errors that could be found in the execution step.
![execute-errors-example.png](/images/execute-errors-example.png)

There're two reasons that caused execution failed:
## 1. Insufficient prepaid gas to execute the transaction
**Solution:** There're two options. 
1) Manually execute the payload at the destination chain via [Axelarscan UI](/dev/gmp/gmp-tracker-recovery/recovery#manually-execute-a-transfer) or [AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery#1-execute-manually).
2) Pay a new gas amount to the Gas Service contract on the source chain via [Axelarscan UI](/dev/gmp/gmp-tracker-recovery/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain) or [AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery#2-increase-gas-payment).

## 2. Error in the destination contract logic
**What to do next:** We suggest debugging your contract and then making a new call.

<Callout emoji="ℹ️">
  The displayed error is extracted from the data returned by the [Ethers.js](https://github.com/ethers-io/ethers.js/) library, from the data fields `error.error.code` and `error.error.reason`. The displayed error code can be clicked to link to the description of each error code in [Ethers.js's official document](https://docs.ethers.io/v5/api/utils/logger/#errors-ethereum). 
</Callout>