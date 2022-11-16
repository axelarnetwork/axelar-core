# axelard tx distribution

Distribution transactions subcommands

```
axelard tx distribution [flags]
```

## Options

```
  -h, --help   help for distribution
```

## Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

## SEE ALSO

- [axelard tx](/cli-docs/v0_27_0/axelard_tx) - Transactions subcommands
- [axelard tx distribution fund-community-pool](/cli-docs/v0_27_0/axelard_tx_distribution_fund-community-pool) - Funds the community pool with the specified amount
- [axelard tx distribution set-withdraw-addr](/cli-docs/v0_27_0/axelard_tx_distribution_set-withdraw-addr) - change the default withdraw address for rewards associated with an address
- [axelard tx distribution withdraw-all-rewards](/cli-docs/v0_27_0/axelard_tx_distribution_withdraw-all-rewards) - withdraw all delegations rewards for a delegator
- [axelard tx distribution withdraw-rewards](/cli-docs/v0_27_0/axelard_tx_distribution_withdraw-rewards) - Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator
