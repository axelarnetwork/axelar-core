## axelard query ibc channel unreceived-packets

Query all the unreceived packets associated with a channel

### Synopsis

Determine if a packet, given a list of packet commitment sequences, is unreceived.

The return value represents:

- Unreceived packet commitments: no acknowledgement exists on receiving chain for the given packet commitment sequence on sending chain.

```
axelard query ibc channel unreceived-packets [port-id] [channel-id] [flags]
```

### Examples

```
<appd> query ibc channel unreceived-packets [port-id] [channel-id] --sequences=1,2,3
```

### Options

```
      --height int             Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                   help for unreceived-packets
      --node string            <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --sequences int64Slice   comma separated list of packet sequence numbers (default [])
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
