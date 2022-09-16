# Refund the prepaid gas

The prepaid gas to `payGasForContractCall` or `payGasForContractCallWithToken` could exceed the actual amount needed for relaying a message to the destination contract.

The Executor Service automatically tracks the excess gas amount and refund it to the payer's wallet address by calling `Refund` in the Gas Service contract.
```
The refunded amount = The total paid amount - the network base fee - the actual gas used - the estimated gas for transferring the refund.
```
You can check the refund status on Axelarscan UI. See [Monitoring state of GMP transactions](/dev/monitor-recover/monitoring#1-axelarscan-ui).

![gmp-gas-refund.png](/images/gmp-gas-refund.png)