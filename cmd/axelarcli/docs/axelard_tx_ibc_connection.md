## axelard tx ibc connection

IBC connection transaction subcommands

```
axelard tx ibc connection [flags]
```

### Options

```
  -h, --help   help for connection
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

- [axelard tx ibc](axelard_tx_ibc.md)	 - IBC transaction subcommands
- [axelard tx ibc connection open-ack](axelard_tx_ibc_connection_open-ack.md)	 - relay the acceptance of a connection open attempt
- [axelard tx ibc connection open-confirm](axelard_tx_ibc_connection_open-confirm.md)	 - confirm to chain B that connection is open on chain A
- [axelard tx ibc connection open-init](axelard_tx_ibc_connection_open-init.md)	 - Initialize connection on chain A
- [axelard tx ibc connection open-try](axelard_tx_ibc_connection_open-try.md)	 - initiate connection handshake between two chains
