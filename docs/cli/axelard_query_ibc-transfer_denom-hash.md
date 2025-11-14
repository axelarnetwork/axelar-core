## axelard query ibc-transfer denom-hash

Query the denom hash info from a given denom trace

### Synopsis

Query the denom hash info from a given denom trace

```
axelard query ibc-transfer denom-hash [trace] [flags]
```

### Examples

```
axelard query ibc-transfer denom-hash transfer/channel-0/uatom
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for denom-hash
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query ibc-transfer](axelard_query_ibc-transfer.md)	 - IBC fungible token transfer query subcommands

