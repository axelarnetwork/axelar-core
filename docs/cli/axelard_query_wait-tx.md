## axelard query wait-tx

Wait for a transaction to be included in a block

### Synopsis

Subscribes to a CometBFT WebSocket connection and waits for a transaction event with the given hash.

```
axelard query wait-tx [hash] [flags]
```

### Examples

```
By providing the transaction hash:
$ axelard q wait-tx [hash]

Or, by piping a "tx" command:
$ axelard tx [flags] | axelard q wait-tx

```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for wait-tx
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
      --timeout duration   The maximum time to wait for the transaction to be included in a block (default 15s)
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query](axelard_query.md) - Querying subcommands
