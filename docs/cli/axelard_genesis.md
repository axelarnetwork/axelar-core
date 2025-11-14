## axelard genesis

Application's genesis-related subcommands

```
axelard genesis [flags]
```

### Options

```
  -h, --help   help for genesis
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

* [axelard](axelard.md)	 - Axelar App
* [axelard genesis add-genesis-account](axelard_genesis_add-genesis-account.md)	 - Add a genesis account to genesis.json
* [axelard genesis bulk-add-genesis-account](axelard_genesis_bulk-add-genesis-account.md)	 - Bulk add genesis accounts to genesis.json
* [axelard genesis collect-gentxs](axelard_genesis_collect-gentxs.md)	 - Collect genesis txs and output a genesis.json file
* [axelard genesis gentx](axelard_genesis_gentx.md)	 - Generate a genesis tx carrying a self delegation
* [axelard genesis migrate](axelard_genesis_migrate.md)	 - Migrate genesis to a specified target version
* [axelard genesis validate](axelard_genesis_validate.md)	 - Validates the genesis file at the default location or at the location passed as an arg

