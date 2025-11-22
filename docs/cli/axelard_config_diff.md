## axelard config diff

Outputs all config values that are different from the app.toml defaults.

```
axelard config diff [target-version] <app-toml-path> [flags]
```

### Options

```
  -h, --help   help for diff
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

- [axelard config](axelard_config.md) - Utilities for managing application configuration
