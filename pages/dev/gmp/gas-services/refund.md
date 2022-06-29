# Refund the prepaid gas

Sometimes, the prepaid gas amount to the payGasForContractCall or payGasForContractCallWithToken excesses the actual amount needed for relaying the call to the destination contract. 

Our relayer service automatically tracks the excess gas submitted and refunds it to the payer account. To do so, the executor calls the `Refund` method in the Gas Service contract.

The Refund amount is calculated as follows:
```
Refund amount = The prepaid amount - the actual gas used - the estimated gas for transferring the refund
```
The refund status will be shown up on Axelarscan after the call was already executed to the destination contract.

![gmp-gas-refund.png](/images/gmp-gas-refund.png)