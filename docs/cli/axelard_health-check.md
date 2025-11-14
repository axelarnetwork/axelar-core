## axelard health-check



```
axelard health-check [flags]
```

### Options

```
      --grpc-addr string       the gRPC endpoint to use for this chain
      --grpc-insecure          allow gRPC over insecure channels, if not the server must use TLS
      --height int             Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                   help for health-check
      --node string            <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
      --operator-addr string   operator address
  -o, --output string          Output format (text|json) (default "text")
      --skip-broadcaster       skip broadcaster check
      --skip-operator          skip operator check
      --skip-tofnd             skip tofnd check
      --tofnd-host string      host name for tss daemon (default "localhost")
      --tofnd-port string      port for tss daemon (default "50051")
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

