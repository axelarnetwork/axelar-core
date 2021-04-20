## axelard query tendermint-validator-set

Get the full tendermint validator set at given height

```
axelard query tendermint-validator-set [height] [flags]
```

### Options

```
  -h, --help                     help for tendermint-validator-set
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test) (default "test")
      --limit int                Query number of results returned per page (default 100)
  -n, --node string              Node to connect to (default "tcp://localhost:26657")
      --page int                 Query a specific page of paginated results (default 1)
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query](axelard_query.md)	 - Querying subcommands
