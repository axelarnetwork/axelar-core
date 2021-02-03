## axelarcli query tendermint-validator-set

Get the full tendermint validator set at given height

### Synopsis

Get the full tendermint validator set at given height

```
axelarcli query tendermint-validator-set [height] [flags]
```

### Options

```
  -h, --help          help for tendermint-validator-set
      --indent        indent JSON response
      --limit int     Query number of results returned per page (default 100)
  -n, --node string   Node to connect to (default "tcp://localhost:26657")
      --page int      Query a specific page of paginated results
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli query](axelarcli_query.md)	 - Querying subcommands
