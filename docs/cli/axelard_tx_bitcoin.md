## axelard tx bitcoin

bitcoin transactions subcommands

```
axelard tx bitcoin [flags]
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

- [axelard tx](axelard_tx.md)	 - Transactions subcommands
- [axelard tx bitcoin confirmTxOut](axelard_tx_bitcoin_confirmTxOut.md)	 - Confirm a Bitcoin transaction
- [axelard tx bitcoin link](axelard_tx_bitcoin_link.md)	 - Link a cross chain address to a bitcoin address created by Axelar
- [axelard tx bitcoin register-external-key](axelard_tx_bitcoin_register-external-key.md)	 - Register the external key for bitcoin
- [axelard tx bitcoin sign-master-consolidation](axelard_tx_bitcoin_sign-master-consolidation.md)	 - Create a Bitcoin transaction for consolidating master key UTXOs, and send the change to an address controlled by \[keyID\]
- [axelard tx bitcoin sign-pending-transfers](axelard_tx_bitcoin_sign-pending-transfers.md)	 - Create a Bitcoin transaction for all pending transfers and sign it
