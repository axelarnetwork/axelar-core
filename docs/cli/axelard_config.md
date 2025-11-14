## axelard config

Utilities for managing application configuration

### Options

```
  -h, --help   help for config
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

- [axelard](axelard.md) - Axelar App
- [axelard config diff](axelard_config_diff.md) - Outputs all config values that are different from the app.toml defaults.
- [axelard config get](axelard_config_get.md) - Get an application config value
- [axelard config home](axelard_config_home.md) - Outputs the folder used as the binary home. No home directory is set when using the `confix` tool standalone.
- [axelard config migrate](axelard_config_migrate.md) - Migrate Cosmos SDK app configuration file to the specified version
- [axelard config set](axelard_config_set.md) - Set an application config value
- [axelard config view](axelard_config_view.md) - View the config file
