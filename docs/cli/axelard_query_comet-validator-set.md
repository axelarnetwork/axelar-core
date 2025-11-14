## axelard query comet-validator-set

Get the full CometBFT validator set at given height

```
axelard query comet-validator-set [height] [flags]
```

### Options

```
  -h, --help            help for comet-validator-set
      --limit int       Query number of results returned per page (default 100)
      --node string     <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
      --page int        Query a specific page of paginated results (default 1)
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query](axelard_query.md)	 - Querying subcommands

