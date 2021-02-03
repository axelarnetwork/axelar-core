## axelarcli query distribution validator-outstanding-rewards

Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations

### Synopsis

Query distribution outstanding (un-withdrawn) rewards
for a validator and all their delegations.

Example:
$ <appcli> query distribution validator-outstanding-rewards cosmosvaloper1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne

```
axelarcli query distribution validator-outstanding-rewards [validator] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for validator-outstanding-rewards
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
