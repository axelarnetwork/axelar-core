## axelard tx permission

permission transactions subcommands

```
axelard tx permission [flags]
```

### Options

```
  -h, --help   help for permission
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
- [axelard tx permission deregister-controller](axelard_tx_permission_deregister-controller.md) - Deregister controller account
- [axelard tx permission halt-chain](axelard_tx_permission_halt-chain.md) - Schedule an upgrade plan at the next block height so that every node stops before executing it
- [axelard tx permission pause](axelard_tx_permission_pause.md) - Pause the chain: only the messages allowlisted for the pause level are accepted until it is unpaused
- [axelard tx permission register-controller](axelard_tx_permission_register-controller.md) - Register controller account
- [axelard tx permission unpause](axelard_tx_permission_unpause.md) - Lift the emergency pause
- [axelard tx permission update-governance-key](axelard_tx_permission_update-governance-key.md) - Update the multisig governance key for axelar network
