## axelarcli query supply total

Query the total supply of coins of the chain

### Synopsis

Query total supply of coins that are held by accounts in the
chain.

Example:
$ <appcli> query supply total

To query for the total supply of a specific coin denomination use:
$ <appcli> query supply total stake

```
axelarcli query supply total [denom] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for total
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

- [axelarcli query supply](axelarcli_query_supply.md)	 - Querying commands for the supply module
