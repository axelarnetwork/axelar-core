## axelard query bitcoin rawTx

Returns the encoded hex string of a fully signed transfer and consolidation transaction

```
axelard query bitcoin rawTx [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for rawTx
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query bitcoin](axelard_query_bitcoin.md)	 - bitcoin query subcommands
