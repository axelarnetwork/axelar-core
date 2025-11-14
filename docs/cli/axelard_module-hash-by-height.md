## axelard module-hash-by-height

Get module hashes at a given height

### Synopsis

Get module hashes at a given height. This command is useful for debugging and verifying the state of the application at a given height. Daemon should not be running when calling this command.

```
axelard module-hash-by-height [height] [flags]
```

### Examples

```
axelard module-hash-by-height 16841115
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for module-hash-by-height
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

* [axelard](axelard.md)	 - Axelar App

