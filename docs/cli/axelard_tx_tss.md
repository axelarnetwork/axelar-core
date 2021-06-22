## axelard tx tss

tss transactions subcommands

```
axelard tx tss [flags]
```

### Options

```
  -h, --help   help for tss
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
- [axelard tx tss assign-next](axelard_tx_tss_assign-next.md)	 - Assigns a previously created key with \[keyID\] as the next key for \[chain\]
- [axelard tx tss rotate](axelard_tx_tss_rotate.md)	 - Rotate the given chain from the old key to the previously assigned one
- [axelard tx tss start-keygen](axelard_tx_tss_start-keygen.md)	 - Initiate threshold key generation protocol
