# Transaction Recovery
Occasionally, transactions can get "stuck" in the pipeline from a source to destination chain (e.g. due to one-off issues that arise with relayers that operate on top of the network).

Transactions have typically gotten "stuck" in the pipeline due to:
(A) The transaction failing to relay from the source chain into the Axelar network for processing.
(B) The transaction failing to get executed on the destination chain.

Users can recover a transaction that gets stuck in the pipeline by either: 
1. Searching for the transaction in the Axelarscan UI and having it invoke recovery.
2. Incorporating the [AxelarJS SDK](/dev/axelarjs-sdk/token-transfer-dep-addr) and invoking those methods directly.

## 1. Axelarscan UI

(A) If the transaction failed to relay from the source chain into the Axelar network for processing - (A) above - Axelarscan UI will show that the transaction has not been `APPROVED`. It will show an option to "APPROVE" that will manually resubmit a request to the network.

![gmp-approve.png](/images/gmp-approve.png)

The CALL APPROVED status will be updated once the network approves the transaction.

![gmp-approve-successful.png](/images/gmp-approve-successful.png)

(B) If the transaction failed to get executed on the destination chain, then Axelarscan will provide the option for you to either:
1. Manually execute a transfer on the destination chain, OR
2. Increase gas payment to the gas receiver on the source chain.

### Manually execute a transfer
Click the ‘Connect’ button under the label ‘Execute at destination chain’. Then click the `Execute` button. It triggers the executor service to execute the transaction using the new gas paid at the destination chain. You can check the latest execution result in the `Executed` section.

![gmp-execute.png](/images/gmp-execute.png)

Suppose the manual execution fails, you will get an error message with explained reason.
![gmp-execute-error-reverted.png](/images/gmp-execute-error-reverted.png)

**What to do next:** We suggest debugging your contract and then making a new call. You can try to follow the [Debugging your smart contract](/dev/monitor-recover/recovery) guide.

### Increase gas payment to the gas receiver on the source chain
The prepaid gas to the Gas Service contract could be insufficient when the destination chain is too busy (with many transfers or other conditions). Therefore, Axelarscan provides an option to increase gas payment to relay the transaction. 

To do this:
1. Click the `Connect` button to connect your MetaMask wallet. Then switch the wallet network to the transfer’s source chain by clicking the `Switch Network` button under the label ‘Add gas at source chain‘.
![egmp-pay-gas-button.png](/images/gmp-pay-gas-button.png)
2. Click the `Add gas` button. The new paid gas information will be updated in the `GAS PAID` section. Then, the call will be relayed and executed.
![gmp-pay-gas-success.png](/images/gmp-pay-gas-success.png)

## 2. the AxelarJS SDK

All of the recovery methods above can be done programmatically through our SDK. The benefit of this would be if you would like to incorporate these recovery features above in your application directly. In fact, Axelarscan makes use of all of these methods written into the SDK. 

See SDK docs for [the full transaction recovery API](/dev/axelarjs-sdk/tx-status-query-recovery#query-and-recover-gmp-transactions).
