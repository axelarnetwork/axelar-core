# Axelar gas & executor services

## Executor Service 

Axelar network provides an optional relayer service, called the Executor Service, which observes the gas-paid amount to the Gas Service contract (below). Then, it automatically uses those amounts to relay the approved message to the application’s destination contract.

Users are required to pay gas to the Gas Receiver contract to activate the Executor Service. After an attempt to relay the message to the destination contract, Axelar gas services calculate the remaining gas amount and [refunds](./refund) it to the payer account. The execution result can be monitored on the [Axelarscan](https://axelarscan.io) UI or requested through the AxelarJS SDK. Please see the section titled [Monitoring state of GMP transactions](../monitor-recover/monitoring) for more information.

## Gas Receiver
For any General Message Passing (GMP) transaction, the Axelar network routes transactions to their destination chains. Execution is the final step of the pipeline to the specified destination contract address on the destination chain. It is invoked in one of two ways.

1. Manually paid by the user/application on the destination chain.
2. Executed automatically by Axelar if the user/application prepays gas to the Gas Receiver contract on the source chain.

Gas Receiver is a smart contract deployed on every EVM that is provided by Axelar. Our gas services provide users/applications the ability to prepay the full estimated cost for any GMP transaction (from source to destination chains) with the convenience of a single payment on the source chain, relying on Axelar's relay services to manage the full pipeline. Once gas is paid to Gas Receiver for a GMP transaction, Axelar's relayer services pick up the payment and automatically execute the final General Message Passing call.

Developers can use Axelar gas services by prepaying upfront the relayer gas fee on the source chain, thereby covering the cost of gas to execute the final transaction on the destination chain.