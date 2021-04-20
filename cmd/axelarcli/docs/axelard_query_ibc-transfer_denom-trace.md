## axelard query ibc-transfer denom-trace

Query the denom trace info from a given trace hash

### Synopsis

Query the denom trace info from a given trace hash

```
axelard query ibc-transfer denom-trace [hash] [flags]
```

### Examples

```
<appd> query ibc-transfer denom-trace [hash]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for denom-trace
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query ibc-transfer](axelard_query_ibc-transfer.md)	 - IBC fungible token transfer query subcommands
