## axelard query ibc channel end

Query a channel end

### Synopsis

Query an IBC channel end from a port and channel identifiers

```
axelard query ibc channel end [port-id] [channel-id] [flags]
```

### Examples

```
<appd> query ibc channel end [port-id] [channel-id]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for end
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
      --prove           show proofs for the query results (default true)
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query ibc channel](axelard_query_ibc_channel.md)	 - IBC channel query subcommands
