## axelarcli query slashing params

Query the current slashing parameters

### Synopsis

Query genesis parameters for the slashing module:

$ <appcli> query slashing params

```
axelarcli query slashing params [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for params
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

- [axelarcli query slashing](axelarcli_query_slashing.md)	 - Querying commands for the slashing module
