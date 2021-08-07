## axelard set-genesis-tss

Set the genesis parameters for the tss module

```
axelard set-genesis-tss [flags]
```

### Options

```
  -h, --help                      help for set-genesis-tss
      --keygen-master string      The minimum % of stake that must be online to authorize generation of a new master key in the system (e.g., "9/10").
      --keygen-secondary string   The minimum % of stake that must be online to authorize generation of a new master key in the system (e.g., "9/10").
      --locking-period int        A positive integer representing the locking period for validators in terms of number of blocks
      --safety-master string      The safety threshold with which Axelar Core will run the keygen protocol for a master key (e.g., "2/3").
      --safety-secondary string   The safety threshold with which Axelar Core will run the keygen protocol for a master key (e.g., "2/3").
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

- [axelard](axelard.md)	 - Axelar App
