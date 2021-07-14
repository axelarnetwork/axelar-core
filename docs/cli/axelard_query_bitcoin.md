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
- [axelard query bitcoin consolidationTxState](axelard_query_bitcoin_consolidationTxState.md)	 - Returns the state of the consolidation transaction as seen by Axelar network
- [axelard query bitcoin deposit-address](axelard_query_bitcoin_deposit-address.md)	 - Returns a bitcoin deposit address for a recipient address on another blockchain
- [axelard query bitcoin master-address](axelard_query_bitcoin_master-address.md)	 - Returns the bitcoin address of the current master key, and optionally the key's ID
- [axelard query bitcoin minWithdraw](axelard_query_bitcoin_minWithdraw.md)	 - Returns the minimum withdraw amount in satoshi
- [axelard query bitcoin nextMasterKeyID](axelard_query_bitcoin_nextMasterKeyID.md)	 - Returns the next assigned master key ID
- [axelard query bitcoin rawPayForConsolidationTx](axelard_query_bitcoin_rawPayForConsolidationTx.md)	 - Returns the encoded hex string of a fully signed transaction that pays for the consolidation transaction
- [axelard query bitcoin rawTx](axelard_query_bitcoin_rawTx.md)	 - Returns the encoded hex string of a fully signed transfer and consolidation transaction
- [axelard query bitcoin txState](axelard_query_bitcoin_txState.md)	 - Returns the state of a bitcoin transaction as seen by Axelar network
