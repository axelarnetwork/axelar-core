## axelarcli query staking delegations

Query all delegations made by one delegator

### Synopsis

Query delegations for an individual delegator on all validators.

Example:
$ <appcli> query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p

```
axelarcli query staking delegations [delegator-addr] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for delegations
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
