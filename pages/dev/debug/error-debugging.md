# Execution error messages

import Callout from 'nextra-theme-docs/callout'

Below are some possible errors that could be found in the execution step.
![execute-errors-example.png](/images/execute-errors-example.png)

<Callout emoji="ℹ️">
  The displayed error is extracted from the data returned by the [Ethers.js](https://github.com/ethers-io/ethers.js/) library, from the data fields `error.error.code` and `error.error.reason`. The displayed error code can be clicked to link to the description of each error code in [Ethers.js's official document](https://docs.ethers.io/v5/api/utils/logger/#errors-ethereum).
</Callout>

There are two reasons that executions fail:

### 1. Insufficient prepaid gas to execute the transaction

There are two options to solve this problem.

1. Manually execute the payload at the destination chain via [Axelarscan UI](../monitor-recover/recovery#manually-execute-a-transfer) or [AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery#1-execute-manually).
2. Pay a new gas amount to the Gas Receiver contract on the source chain via [Axelarscan UI](../monitor-recover/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain) or [AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery#2-increase-gas-payment).

### 2. Error in the destination contract logic

**What to do next:** We suggest debugging your contract and then making a new call. You can try to follow the [Debugging your smart contract](..//debug/debugging-your-smart-contract) guide.
