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
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard tx](axelard_tx.md) - Transactions subcommands
- [axelard tx distribution community-pool-spend-proposal](axelard_tx_distribution_community-pool-spend-proposal.md) - Submit a proposal to spend from the community pool
- [axelard tx distribution fund-community-pool](axelard_tx_distribution_fund-community-pool.md) - Funds the community pool with the specified amount
- [axelard tx distribution fund-validator-rewards-pool](axelard_tx_distribution_fund-validator-rewards-pool.md) - Fund the validator rewards pool with the specified amount
- [axelard tx distribution set-withdraw-addr](axelard_tx_distribution_set-withdraw-addr.md) - change the default withdraw address for rewards associated with an address
- [axelard tx distribution update-params-proposal](axelard_tx_distribution_update-params-proposal.md) - Submit a proposal to update distribution module params. Note: the entire params must be provided.
- [axelard tx distribution withdraw-all-rewards](axelard_tx_distribution_withdraw-all-rewards.md) - withdraw all delegations rewards for a delegator
- [axelard tx distribution withdraw-rewards](axelard_tx_distribution_withdraw-rewards.md) - Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator
- [axelard tx distribution withdraw-validator-commission](axelard_tx_distribution_withdraw-validator-commission.md) - Withdraw commissions from a validator address (must be a validator operator)
