## axelard config set

Set an application config value

### Synopsis

Set an application config value. The [config] argument must be the path of the file when using the `confix` tool standalone, otherwise it must be the name of the config file without the .toml extension.

```
axelard config set [config] [key] [value] [flags]
```

### Options

```
  -h, --help            help for set
  -s, --skip-validate   skip configuration validation (allows to mutate unknown configurations)
      --stdout          print the updated config to stdout
  -v, --verbose         log changes to stderr
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

