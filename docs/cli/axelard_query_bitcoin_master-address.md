## axelard query bitcoin master-address

Returns the bitcoin address of the current master key, and optionally the key's ID

```
axelard query bitcoin master-address [flags]
```

### Options

```
      --height int       Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help             help for master-address
      --include-key-id   include the current master key ID in the output
      --node string      <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
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

- [axelard query bitcoin](axelard_query_bitcoin.md)	 - bitcoin query subcommands
