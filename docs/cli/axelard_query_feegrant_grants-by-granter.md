## axelard query feegrant grants-by-granter

Query all grants by a granter

### Synopsis

Queries all the grants issued for a granter address.

Example:
$ <appd> query feegrant grants-by-granter \[granter\]

```
axelard query feegrant grants-by-granter [granter] [flags]
```

### Options

```
      --count-total       count total number of records in grants to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for grants-by-granter
      --limit uint        pagination limit of grants to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of grants to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of grants to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of grants to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query feegrant](axelard_query_feegrant.md)	 - Querying commands for the feegrant module
