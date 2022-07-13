# Refund the prepaid gas

The gas amount paid to the `payGasForContractCall` or `payGasForContractCallWithToken` could exceed the actual amount needed for relaying the message to the destination contract. 

Our executor service automatically tracks the excess gas submitted and refunds it to the payer account. To do so, the service calls `Refund` method in the Gas Service contract.

The Refund amount is calculated as follows:
```
Refund amount = The prepaid amount - the actual gas used - the estimated gas for transferring the refund
```
The refund status will be shown up on Axelarscan after the call was already executed to the destination contract.

![gmp-gas-refund.png](/images/gmp-gas-refund.png)