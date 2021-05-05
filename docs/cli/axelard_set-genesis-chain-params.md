## axelard set-genesis-chain-params

Set the chain's parameters in genesis.json

### Synopsis

Set the chain's parameters in genesis.json. The provided chain must be one of those axelar supports.

```
axelard set-genesis-chain-params [chain] [flags]
```

### Options

```
      --confirmation-height uint   Confirmation height to set for the given chain.
  -h, --help                       help for set-genesis-chain-params
      --network string             Name of the network to set for the given chain.
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
