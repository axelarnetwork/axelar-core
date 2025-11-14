## axelard init

Initialize private validator, p2p, genesis, and application configuration files

### Synopsis

Initialize validators's and node's configuration files.

```
axelard init [moniker] [flags]
```

### Options

```
      --chain-id string        genesis file chain-id, if left blank will be randomly created (default "axelar")
      --default-denom string   genesis file default denomination, if left blank default value is 'stake'
  -h, --help                   help for init
      --home string            node's home directory (default "$HOME/.axelar")
      --initial-height int     specify the initial block height at genesis (default 1)
  -o, --overwrite              overwrite the genesis.json file
      --recover                provide seed phrase to recover existing key instead of creating
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

