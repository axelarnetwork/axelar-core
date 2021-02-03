## axelarcli query staking unbonding-delegations

Query all unbonding-delegations records for one delegator

### Synopsis

Query unbonding delegations for an individual delegator.

Example:
$ <appcli> query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p

```
axelarcli query staking unbonding-delegations [delegator-addr] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for unbonding-delegations
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
