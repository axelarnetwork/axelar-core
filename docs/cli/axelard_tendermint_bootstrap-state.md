## axelard tendermint bootstrap-state

Bootstrap CometBFT state at an arbitrary block height using a light client

```
axelard tendermint bootstrap-state [flags]
```

### Options

```
      --height int   Block height to bootstrap state at, if not provided it uses the latest block height in app state
  -h, --help         help for bootstrap-state
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

- [axelard tendermint](axelard_tendermint.md) - Tendermint subcommands
