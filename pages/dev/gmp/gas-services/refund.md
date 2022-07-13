# Refund the prepaid gas

The prepaid gas to `payGasForContractCall` or `payGasForContractCallWithToken` could exceed the needed amount for relaying a message to the destination contract.

The [Executor service](/dev/gmp/executor-service) automatically calculates the excess gas submitted and determines the amount to refund.
```
The refunded amount = The prepaid amount - the actual gas used - the estimated gas for transferring the refund
```
After getting the refund amount, the service calls `Refund` in the Gas Service contract to refund it to the payer account. Then, the refund status will be shown up on Axelarscan UI.

![gmp-gas-refund.png](/images/gmp-gas-refund.png)