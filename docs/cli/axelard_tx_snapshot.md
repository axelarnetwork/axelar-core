## axelard tx snapshot

snapshot transactions subcommands

```
axelard tx snapshot [flags]
```

### Options

```
  -h, --help   help for snapshot
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

- [axelard tx](axelard_tx.md)	 - Transactions subcommands
- [axelard tx snapshot deactivate-proxy](axelard_tx_snapshot_deactivate-proxy.md)	 - Deactivate the proxy account of the sender
- [axelard tx snapshot proxy-ready](axelard_tx_snapshot_proxy-ready.md)	 - Establish a proxy as ready to be registered by the specified operator address
- [axelard tx snapshot register-proxy](axelard_tx_snapshot_register-proxy.md)	 - Register a proxy account for a specific validator principal to broadcast transactions in its stead
- [axelard tx snapshot send-tokens](axelard_tx_snapshot_send-tokens.md)	 - Sends the specified amount of tokens to the designated addresses
