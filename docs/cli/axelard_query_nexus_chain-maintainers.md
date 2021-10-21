## axelard query nexus chain-maintainers

Returns the chain maintainers for the given chain

```
axelard query nexus chain-maintainers [chain] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for chain-maintainers
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query nexus](axelard_query_nexus.md)	 - Querying commands for the nexus module
