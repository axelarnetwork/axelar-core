# axelard set-genesis-staking

Set the genesis parameters for the staking module

```
axelard set-genesis-staking [flags]
```

## Options

```
      --bond-denom string           A string representing bondable coin denomination
  -h, --help                        help for set-genesis-staking
      --historical-entries uint32   A positive integer representing the number of historical entries kept
      --max-entries uint32          A positive integer representing the maximum number of redelegations allowed within the unbonding period
      --max-validators uint32       A positive integer representing the maximum number of validators (max uint16 = 65535)
      --unbonding-period string     Time duration of unbonding (e.g., "6h").
```

## Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

## SEE ALSO

- [axelard](/cli-docs/v0_27_0/axelard) - Axelar App
