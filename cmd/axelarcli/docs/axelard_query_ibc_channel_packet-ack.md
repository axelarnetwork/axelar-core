## axelard query ibc channel packet-ack

Query a packet acknowledgement

### Synopsis

Query a packet acknowledgement

```
axelard query ibc channel packet-ack [port-id] [channel-id] [sequence] [flags]
```

### Examples

```
<appd> query ibc channel packet-ack [port-id] [channel-id] [sequence]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for packet-ack
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
