## axelard set-genesis-gov

Set the genesis parameters for the governance module

```
axelard set-genesis-gov [flags]
```

### Options

```
  -h, --help                      help for set-genesis-gov
      --max-deposit-period uint   Maximum period for AXL holders to deposit on a proposal (time ns)
      --minimum-deposit string    Minimum deposit for a proposal to enter voting period
      --voting-period uint        Length of the voting period (time ns)
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
