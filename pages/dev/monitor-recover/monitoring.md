# Monitoring state of GMP transactions

Axelar provides two options to check each GMP transaction status: 
1. The Axelarscan UI. 
2. The [AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery).

## 1. Axelarscan UI
Anyone can view General Message Passing calls in realtime on Axelarscan: [Mainnet](https://axelarscan.io/gmp) | [Testnet](https://testnet.axelarscan.io/gmp).

![gmp-tracker.png](/images/gmp-tracker.png)

You can also search for a particular transfer by a transaction hash or a sender address via the search bar. 
![gmp-searchbar.png](/images/gmp-searchbar.png)

Each GMP call comprises five statuses, as described below.
![gmp-detailed-page .png](/images/gmp-detailed-page.png)

- **CONTRACT CALL** provides the contract call (`callContract` or `callContractWithToken`) information, including the transaction hash, the block height on the source chain, the gateway address, etc.
- **GAS PAID** displays the information of gas prepaid to Axelar Gas Receiver contract.
- **CALL APPROVED** displays the information of the call approval. This section will be updated once the GMP call is approved by the Axelar network.
- **EXECUTED** informs the executed result whether it is successful or not. If you see an error message in this section, we suggest following this [guide](/dev/debug/error-debugging) to find the root cause and recover the transfer.
- **GAS REFUNDED** provides the refund information (if any), including the amount of gas paid, the amount of gas used, the refund amount, etc.

import Callout from 'nextra-theme-docs/callout'

<Callout emoji="ℹ️">
	If the `Insufficient Fee` tag appears, it means that the prepaid gas is not enough to relay the transaction. Please follow the [Increase gas payment to the gas receiver on the source chain](/dev/monitor-recover/recovery#increase-gas-payment-to-the-gas-receiver-on-the-source-chain) section to recover the transfer.
	![error-msg-insufficient-fee.png](/images/error-msg-insufficient-fee.png)
</Callout>

## 2. AxelarJS SDK

See the SDK docs section, [Query transaction status by txHash](/dev/axelarjs-sdk/tx-status-query-recovery#query-transaction-status-by-txhash).
