## axelard query ibc-transfer denom-hash

Query the denom hash info from a given denom trace

### Synopsis

Query the denom hash info from a given denom trace

```
axelard query ibc-transfer denom-hash [trace] [flags]
```

### Examples

```
<appd> query ibc-transfer denom-hash transfer/channel-0/uatom
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for denom-hash
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query ibc-transfer](axelard_query_ibc-transfer.md)	 - IBC fungible token transfer query subcommands
