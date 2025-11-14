## axelard query tx

Query for a transaction by hash, "<addr>/<seq>" combination or comma-separated signatures in a committed block

### Synopsis

Example:
$ axelard query tx <hash>
$ axelard query tx --type=acc_seq <addr>/<sequence>
$ axelard query tx --type=signature <sig1_base64>,<sig2_base64...>

```
axelard query tx --type=[hash|acc_seq|signature] [hash|acc_seq|signature] [flags]
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for tx
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
      --type string        The type to be used when querying tx, can be one of "hash", "acc_seq", "signature" (default "hash")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query](axelard_query.md)	 - Querying subcommands

