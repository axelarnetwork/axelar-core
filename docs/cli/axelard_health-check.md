## axelard health-check

```
axelard health-check [flags]
```

### Options

```
      --broadcaster-addr string   broadcaster address
      --check-broadcaster         assert that broadcaster has funds (requires --broadcaster-addr) (default true)
      --check-operator            perform healthcheck upon the operator address (requires --broadcaster-addr) (default true)
      --check-tofnd               perform simple tofnd ping (default true)
      --context-timeout string    context timeout for the grpc (default "2h0m0s")
      --height int                Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                      help for health-check
      --node string               <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --tofnd-host string         host name for tss daemon (default "localhost")
      --tofnd-port string         port for tss daemon (default "50051")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
