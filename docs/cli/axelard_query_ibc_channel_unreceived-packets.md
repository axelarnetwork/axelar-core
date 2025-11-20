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
axelard query ibc channel unreceived-packets [port-id] [channel-id] --sequences=1,2,3
```

### Options

```
      --grpc-addr string       the gRPC endpoint to use for this chain
      --grpc-insecure          allow gRPC over insecure channels, if not the server must use TLS
      --height int             Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                   help for unreceived-packets
      --node string            <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string          Output format (text|json) (default "text")
      --sequences int64Slice   comma separated list of packet sequence numbers (default [])
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query ibc channel](axelard_query_ibc_channel.md) - IBC channel query subcommands
