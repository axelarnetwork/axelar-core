## axelarcli query bitcoin sendTransfers

Send a transaction to Bitcoin that consolidates deposits and withdrawals

### Synopsis

Send a transaction to Bitcoin that consolidates deposits and withdrawals

```
axelarcli query bitcoin sendTransfers [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for sendTransfers
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli query bitcoin](axelarcli_query_bitcoin.md)	 - bitcoin query subcommands
