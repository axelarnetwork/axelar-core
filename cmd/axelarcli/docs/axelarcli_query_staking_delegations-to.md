## axelarcli query staking delegations-to

Query all delegations made to one validator

### Synopsis

Query delegations on an individual validator.

Example:
$ <appcli> query staking delegations-to cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
axelarcli query staking delegations-to [validator-addr] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for delegations-to
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

- [axelarcli query staking](axelarcli_query_staking.md)	 - Querying commands for the staking module
