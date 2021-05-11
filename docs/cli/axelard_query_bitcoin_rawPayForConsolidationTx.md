## axelard query bitcoin rawPayForConsolidationTx

Returns the encoded hex string of a fully signed transaction that pays for the consolidation transaction

```
axelard query bitcoin rawPayForConsolidationTx [flags]
```

### Options

```
      --fee-rate int   fee rate to be set for the child-pay-for-parent transaction
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for rawPayForConsolidationTx
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
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

- [axelard query bitcoin](axelard_query_bitcoin.md)	 - bitcoin query subcommands
