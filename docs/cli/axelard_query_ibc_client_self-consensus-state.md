## axelard query ibc client self-consensus-state

Query the self consensus state for this chain

### Synopsis

Query the self consensus state for this chain. This result may be used for verifying IBC clients representing this chain which are hosted on counterparty chains.

```
axelard query ibc client self-consensus-state [flags]
```

### Examples

```
axelard query ibc client self-consensus-state
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for self-consensus-state
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

* [axelard query ibc client](axelard_query_ibc_client.md)	 - IBC client query subcommands

