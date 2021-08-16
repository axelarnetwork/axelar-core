## axelard query bitcoin

bitcoin query subcommands

```
axelard query bitcoin [flags]
```

### Options

```
  -h, --help   help for bitcoin
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

- [axelard query](axelard_query.md)	 - Querying subcommands
- [axelard query bitcoin consolidation-address](axelard_query_bitcoin_consolidation-address.md)	 - Returns the bitcoin consolidation address
- [axelard query bitcoin deposit-address](axelard_query_bitcoin_deposit-address.md)	 - Returns a bitcoin deposit address for a recipient address on another blockchain
- [axelard query bitcoin deposit-status](axelard_query_bitcoin_deposit-status.md)	 - Returns the status of the bitcoin deposit with the given outpoint
- [axelard query bitcoin latest-tx](axelard_query_bitcoin_latest-tx.md)	 - Returns the latest consolidation transaction of the given key role
- [axelard query bitcoin min-output-amount](axelard_query_bitcoin_min-output-amount.md)	 - Returns the minimum amount allowed for any transaction output in satoshi
- [axelard query bitcoin next-key-id](axelard_query_bitcoin_next-key-id.md)	 - Returns the ID of the next assigned key
- [axelard query bitcoin signed-tx](axelard_query_bitcoin_signed-tx.md)	 - Returns the signed consolidation transaction of the given transaction hash
