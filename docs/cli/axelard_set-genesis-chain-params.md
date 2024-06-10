## axelard set-genesis-chain-params

Set chain parameters in genesis.json

### Synopsis

Set chain parameters in genesis.json. The provided platform must be one of those axelar supports (currently only EVM).

```
axelard set-genesis-chain-params evm [chain] [flags]
```

### Options

```
      --confirmation-height uint    Confirmation height to set for the given chain.
      --evm-chain-id string         Integer representing the chain ID (EVM only).
      --evm-network-name string     Network name (EVM only).
  -h, --help                        help for set-genesis-chain-params
      --home string                 node's home directory (default "$HOME/.axelar")
      --network string              Name of the network to set for the given chain.
      --revote-locking-period int   Revote locking period to set for the given chain.
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
