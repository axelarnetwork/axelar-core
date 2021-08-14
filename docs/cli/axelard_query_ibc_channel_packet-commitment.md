## axelard query ibc channel packet-commitment

Query a packet commitment

### Synopsis

Query a packet commitment

```
axelard query ibc channel packet-commitment [port-id] [channel-id] [sequence] [flags]
```

### Examples

```
<appd> query ibc channel packet-commitment [port-id] [channel-id] [sequence]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for packet-commitment
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --prove         show proofs for the query results (default true)
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

- [axelard query ibc channel](axelard_query_ibc_channel.md)	 - IBC channel query subcommands
