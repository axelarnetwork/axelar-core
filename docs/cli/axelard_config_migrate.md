## axelard config migrate

Migrate Cosmos SDK app configuration file to the specified version

### Synopsis

Migrate the contents of the Cosmos SDK app configuration (app.toml) to the specified version.
The output is written in-place unless --stdout is provided.
In case of any error in updating the file, no output is written.

```
axelard config migrate [target-version] <app-toml-path> (options) [flags]
```

### Options

```
  -h, --help            help for migrate
      --skip-validate   skip configuration validation (allows to migrate unknown configurations)
      --stdout          print the updated config to stdout
      --verbose         log changes to stderr
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

