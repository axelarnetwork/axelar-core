# Executor service 

Axelar network provides an optional relayer service, called Executor service, which observes the gas-paid amount to the [Gas Service](/dev/gmp/gas-services/overview) contract, and automatically uses those amounts to relay the approved message to the application’s destination contract.

To activate the use of the executor service, users are required to pay gas to the [Gas Service](/dev/gmp/gas-services/overview) contract on the source chain. The executor service will then handle relaying the message to the destination contract on the destination chain, and also [refund](/dev/gmp/gas-services/refund) the remaining gas used to the payer account at the end of the process. 

So, only a couple of things are required to make a GMP transfer with the Executor service: 1) call the contract (`callContract` or `callContractWithToken`) and 2) pay gas to the [Gas Service](/dev/gmp/gas-services/overview) contract.

## Two-way call
The Executor service supports Two-way call, where a message is sent from a source chain, immediately executed at a destination chain, and sent another message back to the source chain.

The service monitors if there's another contract call immediately executed within the same executed transaction on the destination (to send a message back to the source chain). The service will then automatically uses the remaining pre-paid to relay the second call of the two-way call. 

In case the remaining gas amount (from the first contract call) is insufficient for the second call, the service will refund it to the payer's address. Users still can pay a new gas amount to relay the second call of the transfer through the [Axelar SDK](/dev/axelarjs-sdk/tx-status-query-recovery#22-erc-20-gas-payment) or [Axelarscan UI](http://localhost:3000/dev/gmp/gmp-tracker-recovery/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain).