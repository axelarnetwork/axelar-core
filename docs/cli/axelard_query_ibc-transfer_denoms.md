## axelard query ibc-transfer denoms

Query for all token denominations

### Synopsis

Query for all token denominations

```
axelard query ibc-transfer denoms [flags]
```

### Examples

```
axelard query ibc-transfer denoms
```

### Options

```
      --count-total        count total number of records in denominations to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for denoms
      --limit uint         pagination limit of denominations to query for (default 100)
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint        pagination offset of denominations to query for
  -o, --output string      Output format (text|json) (default "text")
      --page uint          pagination page of denominations to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of denominations to query for
      --reverse            results are sorted in descending order
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query ibc-transfer](axelard_query_ibc-transfer.md) - IBC fungible token transfer query subcommands
