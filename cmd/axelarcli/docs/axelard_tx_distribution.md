## axelard tx distribution

Distribution transactions subcommands

```
axelard tx distribution [flags]
```

### Options

```
  -h, --help   help for distribution
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
- [axelard tx distribution fund-community-pool](axelard_tx_distribution_fund-community-pool.md)	 - Funds the community pool with the specified amount
- [axelard tx distribution set-withdraw-addr](axelard_tx_distribution_set-withdraw-addr.md)	 - change the default withdraw address for rewards associated with an address
- [axelard tx distribution withdraw-all-rewards](axelard_tx_distribution_withdraw-all-rewards.md)	 - withdraw all delegations rewards for a delegator
- [axelard tx distribution withdraw-rewards](axelard_tx_distribution_withdraw-rewards.md)	 - Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator
