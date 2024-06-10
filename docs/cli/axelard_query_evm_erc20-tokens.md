## axelard query evm erc20-tokens

Returns the ERC20 tokens for the given chain

```
axelard query evm erc20-tokens [chain] [flags]
```

### Options

```
      --height int          Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                help for erc20-tokens
      --node string         <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string       Output format (text|json) (default "text")
      --token-type string   the token type [external|internal]
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query evm](axelard_query_evm.md)	 - Querying commands for the evm module
