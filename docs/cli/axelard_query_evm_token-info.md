## axelard query evm token-info

Returns the info of token by either symbol or asset

```
axelard query evm token-info [chain] [flags]
```

### Options

```
      --asset string    lookup token by asset name
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for token-info
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --symbol string   lookup token by symbol
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query evm](axelard_query_evm.md)	 - Querying commands for the evm module
