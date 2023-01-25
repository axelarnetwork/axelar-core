## axelard set-genesis-slashing

Set the genesis parameters for the slashing module

```
axelard set-genesis-slashing [flags]
```

### Options

```
      --downtime-jail-duration string       Jail duration due to downtime (e.g., "600s").
  -h, --help                                help for set-genesis-slashing
      --home string                         node's home directory (default "$HOME/.axelar")
      --min-signed-per-window string        Minimum amount of signed blocks per window (e.g., "0.50").
      --signed-blocks-window uint           Block height window to measure liveness of each validator (e.g., 10000).
      --slash-fraction-double-sign string   Slashing fraction due to double signing (e.g., "0.01").
      --slash-fraction-downtime string      Slashing fraction due to downtime (e.g., "0.0001").
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
