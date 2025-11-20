## axelard rollback

rollback Cosmos SDK and CometBFT state by one height

### Synopsis

A state rollback is performed to recover from an incorrect application state transition,
when CometBFT has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also rolls back to height n - 1. No blocks are removed, so upon
restarting CometBFT the transactions in block n will be re-executed against the
application.

```
axelard rollback [flags]
```

### Options

```
      --hard          remove last block as well as state
  -h, --help          help for rollback
      --home string   The application home directory (default "$HOME/.axelar")
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md) - Axelar App
