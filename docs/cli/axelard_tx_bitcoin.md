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
- [axelard tx bitcoin confirm-tx-out](axelard_tx_bitcoin_confirm-tx-out.md)	 - Confirm a Bitcoin transaction
- [axelard tx bitcoin create-master-tx](axelard_tx_bitcoin_create-master-tx.md)	 - Create a Bitcoin transaction for consolidating master key UTXOs, and send the change to an address controlled by \[keyID\]
- [axelard tx bitcoin create-pending-transfers-tx](axelard_tx_bitcoin_create-pending-transfers-tx.md)	 - Create a Bitcoin transaction for all pending transfers
- [axelard tx bitcoin create-rescue-tx](axelard_tx_bitcoin_create-rescue-tx.md)	 - Create a Bitcoin transaction for rescuing the outpoints that were sent to old keys
- [axelard tx bitcoin link](axelard_tx_bitcoin_link.md)	 - Link a cross chain address to a bitcoin address created by Axelar
- [axelard tx bitcoin sign-tx](axelard_tx_bitcoin_sign-tx.md)	 - Sign a consolidation transaction with the current key of given key role
- [axelard tx bitcoin submit-external-signature](axelard_tx_bitcoin_submit-external-signature.md)	 - Submit a signature of the given external key signing the given sig hash
