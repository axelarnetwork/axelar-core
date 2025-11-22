## axelard query evidence evidence

Query for evidence by hash

```
axelard query evidence evidence [hash] [flags]
```

### Examples

```
axelard query evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
```

### Options

```
      --evidence-hash binary   
      --grpc-addr string       the gRPC endpoint to use for this chain
      --grpc-insecure          allow gRPC over insecure channels, if not the server must use TLS
      --height int             Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                   help for evidence
      --no-indent              Do not indent JSON output
      --node string            <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string          Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query evidence](axelard_query_evidence.md) - Querying commands for the evidence module
