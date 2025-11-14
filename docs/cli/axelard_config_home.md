## axelard config home

Outputs the folder used as the binary home. No home directory is set when using the `confix` tool standalone.

### Synopsis

Outputs the folder used as the binary home. In order to change the home directory path, set the $APPD_HOME environment variable, or use the "--home" flag.

```
axelard config home [flags]
```

### Options

```
  -h, --help   help for home
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

* [axelard config](axelard_config.md)	 - Utilities for managing application configuration

