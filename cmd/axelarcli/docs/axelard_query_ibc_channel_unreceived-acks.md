## axelard query ibc channel unreceived-acks

Query all the unreceived acks associated with a channel

### Synopsis

Given a list of acknowledgement sequences from counterparty, determine if an ack on the counterparty chain has been received on the executing chain.

The return value represents:

- Unreceived packet acknowledgement: packet commitment exists on original sending (executing) chain and ack exists on receiving chain.

```
axelard query ibc channel unreceived-acks [port-id] [channel-id] [flags]
```

### Examples

```
<appd> query ibc channel unreceived-acks [port-id] [channel-id] --sequences=1,2,3
```

### Options

```
      --height int             Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                   help for unreceived-acks
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
