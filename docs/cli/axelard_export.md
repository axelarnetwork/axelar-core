## axelard export

Export state to JSON

```
axelard export [flags]
```

### Options

```
      --for-zero-height              Export state to start at height zero (perform preproccessing)
      --height int                   Export state from a particular height (-1 means latest height) (default -1)
  -h, --help                         help for export
      --home string                  The application home directory (default "$HOME/.axelar")
      --jail-allowed-addrs strings   Comma-separated list of operator addresses of jailed validators to unjail
      --modules-to-export strings    Comma-separated list of modules to export. If empty, will export all modules
      --output-document string       Exported state is written to the given file instead of STDOUT
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard](axelard.md)	 - Axelar App

