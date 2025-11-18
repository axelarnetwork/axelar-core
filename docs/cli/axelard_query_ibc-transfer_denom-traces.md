## axelard query ibc-transfer denom-traces

Query the trace info for all token denominations

### Synopsis

Query the trace info for all token denominations

```
axelard query ibc-transfer denom-traces [flags]
```

### Examples

```
axelard query ibc-transfer denom-traces
```

### Options

```
      --count-total        count total number of records in denominations trace to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for denom-traces
      --limit uint         pagination limit of denominations trace to query for (default 100)
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint        pagination offset of denominations trace to query for
  -o, --output string      Output format (text|json) (default "text")
      --page uint          pagination page of denominations trace to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of denominations trace to query for
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
