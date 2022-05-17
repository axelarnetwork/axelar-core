# General Message Passing status tracker

### View transfers status
You can view General Message Passing transactions on the GMP page of Axelarscan.
- mainnet: https://axelarscan.io/gmp
- testnet: https://testnet.axelarscan.io/gmp.

![gmp-tracker.png](/images/gmp-tracker.png)

To search for a particular transfer, you can enter either a transaction hash or a sender address in the search bar. 
![gmp-searchbar.png](/images/gmp-searchbar.png)

### searchGMP API

This API endpoint allows you to programmatically get the General Message Passing status via an HTTP request.

#### HTTP Request
**Mainnet:** `GET https://api.gmp.axelarscan.io`<br />
**Testnet:** `GET https://testnet.api.gmp.axelarscan.io`

#### Query Parameters
| Parameter          | Type     | Description                                                                                                           |
| ------------------ | -------- | --------------------------------------------------------------------------------------------------------------------- |
| `method`<br />**(required field)** | string   | must insert `searchGMP` as mandatory                                                                                          |
| `txHash`           | string   | the transaction hash on source chain                                                                              |
| `sourceChain`      | string   | the source chain of the transfer                                                                                      |
| `destinationChain` | string   | the destination chain of the transfer                                                                                 |
| `senderAddress`    | string   | the sender address                                                                                                    |
| `sourceAddress`    | string   | the source address of the transfer                                                                                    |
| `contractAddress`  | string   | the destination contract address that the transfer will be executed to                                                                   |
| `event`            | string   | the event emitted on the destination gateway contracts. This can be either `contractCall` or `contractCallWithToken`. |
| `relayerAddress`   | string   | the relayer address used in the transfer                                                                                            |
| `status`           | string   | `approving`: the transfer is waiting for approval at the destination chain<br />`approved`: the transfer is approved at the destination chain and waiting to be executed to the destination contract<br />`executed`: the transfer is successfully executed to the destination contract |
| `fromTime`         | unix time | the start timestamp of the transfer                                                                                     |
| `toTime`           | unix time | the end timestamp of the transfer                                                                                       |
| `from`             | number   | the records' index for pagination                                                                                     |
| `size`             | number   | the number of the returned records. This field is used for pagination.                                                |

### Execute Manually
Axelar provides a relayer service that automatically executes transfers for you. Alternatively, you also can execute transfers manually through the Axelarscan's GMP page by connecting a wallet, and click the Execute button. Please note that you need to pay gas to process the transaction.

![gmp-connect-wallet.png](/images/gmp-connect-wallet.png)
![gmp-execute-button.png](/images/gmp-execute-button.png)