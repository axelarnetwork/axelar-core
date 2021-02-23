## axelarcli query ethereum deploy-gateway

Obtain a raw transaction for the deployment of the Axelar Geteway.

### Synopsis

Obtain a raw transaction for the deployment of the Axelar Geteway.

```
axelarcli query ethereum deploy-gateway [flags]
```

### Options

```
      --gas-limit uint     Ethereum gas limit to use in the transaction (default value is 3000000). Set to 0 to estimate gas limit at the node. (default 3000000)
      --gas-price string   Ethereum gas price to use in the transaction. If falg is omitted (or value set to 0), the gas price will be suggested by the node (default "0")
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for deploy-gateway
      --indent             Add indent to JSON response
      --ledger             Use a connected Ledger device
      --node string        <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node         Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli query ethereum](axelarcli_query_ethereum.md)	 - Querying commands for the ethereum module
