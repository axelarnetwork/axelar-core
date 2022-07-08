# Monitoring State of GMP Transactions

Axelar provides two options to check each GMP transaction status: 
1. The Axelarscan UI. 
2. The [[AxelarJS SDK](/dev/axelarjs-sdk/tx-status-query-recovery)].

## 1. Axelarscan UI
Anyone can view a General Message Passing transaction on the GMP page of the Axelarscan block explorer.
- mainnet: https://axelarscan.io/gmp.
- testnet: https://testnet.axelarscan.io/gmp.

![gmp-tracker.png](/images/gmp-tracker-2.png)

To search for a particular transfer, enter a transaction hash or a sender address in the search bar. 
![gmp-searchbar.png](/images/gmp-searchbar.png)

Once you navigate to the detailed transfer page, you will see four main statuses with an additional status, as displayed in the image below.
![gmp-detailed-page .png](/images/gmp-detailed-page.png)

- **CONTRACT CALL** provides the contract call (`callContract` or `callContractWithToken`) information, including the transaction hash, the block height on the source chain, the gateway address, etc.
- **GAS PAID** displays the information of the prepaid gas to Axelar Gas Service contract.
- **CALL APPROVED** displays the information on the call approval. This section will be updated once the Axelar network approves the call. 
- **EXECUTED** informs the execution result whether the execution is successful or not. If it's unsuccessful, there will be [error message](/dev/gmp/gmp-tracker-recovery/error-debugging) with cause displayed in this section. 
- **GAS REFUNDED** (optional) provides the refund information (if any), including the amount of gas paid, the amount of gas used, the refund amount, etc. This information will be shown up only when there’s a refund.


## 2. AxelarJS SDK

See SDK docs for [[Query transaction status by txHash](/dev/axelarjs-sdk/tx-status-query-recovery#query-transaction-status-by-txhash)].
