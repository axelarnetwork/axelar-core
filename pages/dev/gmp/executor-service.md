# Executor service 

Axelar network provides an optional relayer service, called Executor service, which observes the gas-paid amount to the [Gas Service](/dev/gmp/gas-services/overview) contract. Then, it automatically uses those amounts to relay the approved message to the application’s destination contract.

Users are required to pay gas to the [Gas Service](/dev/gmp/gas-services/overview) contract to activate the executor service. After an attempt to relay the message to the destination contract, the service calculates the remaining gas amount and [refunds](/dev/gmp/gas-services/refund) it to the payer account. The execution result can be monitored on Axelarscan UI or requested through the AxelarJS SDK. Please see the [Monitoring State of GMP Transactions](/dev/gmp/gmp-tracker-recovery/monitoring) section for more information.

## Two-way call
The Executor service supports Two-way call, where a message is sent from a source chain, immediately executed at a destination chain, and sent another message back to the source chain.

The service monitors if there's another contract call immediately executed within the same executed transaction on the destination (to send a message back to the source chain). The service will then automatically uses the remaining pre-paid to relay the second call of the two-way call. 

In case the remaining gas amount (from the first contract call) is insufficient for the second call, the service refunds it to the payer's address. Users still can pay a new gas amount to relay the second call of the transfer through the [Axelar SDK](/dev/axelarjs-sdk/tx-status-query-recovery#2-increase-gas-payment) or [Axelarscan UI](/dev/gmp/gmp-tracker-recovery/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain).

