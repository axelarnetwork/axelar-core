## axelarcli query distribution slashes

Query distribution validator slashes

### Synopsis

Query all slashes of a validator for a given block range.

Example:
$ <appcli> query distribution slashes cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100

```
axelarcli query distribution slashes [validator] [start-height] [end-height] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for slashes
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

- [axelarcli query distribution](axelarcli_query_distribution.md)	 - Querying commands for the distribution module
