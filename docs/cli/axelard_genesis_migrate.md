## axelard genesis migrate

Migrate genesis to a specified target version

### Synopsis

Migrate the source genesis into the target version and print to STDOUT

```
axelard genesis migrate [target-version] [genesis-file] [flags]
```

### Examples

```
axelard genesis migrate v0.47 /path/to/genesis.json --chain-id=cosmoshub-3 --genesis-time=2019-04-22T17:00:00Z
```

### Options

```
      --chain-id string          Override chain_id with this flag (default "axelar")
      --genesis-time string      Override genesis_time with this flag
  -h, --help                     help for migrate
      --output-document string   Exported state is written to the given file instead of STDOUT
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

* [axelard genesis](axelard_genesis.md)	 - Application's genesis-related subcommands

