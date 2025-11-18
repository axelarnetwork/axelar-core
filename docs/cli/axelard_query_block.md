## axelard query block

Query for a committed block by height, hash, or event(s)

### Synopsis

Query for a specific committed block using the CometBFT RPC `block` and `block_by_hash` method

```
axelard query block --type=[height|hash] [height|hash] [flags]
```

### Examples

```
$ axelard query block --type=height <height>
$ axelard query block --type=hash <hash>
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for block
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
      --type string        The type to be used when querying tx, can be one of "height", "hash" (default "hash")
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
