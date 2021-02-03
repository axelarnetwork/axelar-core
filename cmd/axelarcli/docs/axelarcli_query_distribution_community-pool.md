## axelarcli query distribution community-pool

Query the amount of coins in the community pool

### Synopsis

Query all coins in the community pool which is under Governance control.

Example:
$ <appcli> query distribution community-pool

```
axelarcli query distribution community-pool [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for community-pool
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
