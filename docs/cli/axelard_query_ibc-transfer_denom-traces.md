## axelard query ibc-transfer denom-traces

Query the trace info for all token denominations

### Synopsis

Query the trace info for all token denominations

```
axelard query ibc-transfer denom-traces [flags]
```

### Examples

```
<appd> query ibc-transfer denom-traces
```

### Options

```
      --count-total       count total number of records in denominations trace to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for denom-traces
      --limit uint        pagination limit of denominations trace to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of denominations trace to query for
      --page uint         pagination page of denominations trace to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of denominations trace to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query ibc-transfer](axelard_query_ibc-transfer.md)	 - IBC fungible token transfer query subcommands
