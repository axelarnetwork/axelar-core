## axelard query bitcoin deposit-addr

Returns a bitcoin deposit address for a recipient address on another blockchain

```
axelard query bitcoin deposit-addr [chain] [recipient address] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for deposit-addr
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query bitcoin](axelard_query_bitcoin.md)	 - bitcoin query subcommands
