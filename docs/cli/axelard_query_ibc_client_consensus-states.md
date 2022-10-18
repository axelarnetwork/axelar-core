## axelard query ibc client consensus-states

Query all the consensus states of a client.

### Synopsis

Query all the consensus states from a given client state.

```
axelard query ibc client consensus-states [client-id] [flags]
```

### Examples

```
<appd> query ibc client consensus-states [client-id]
```

### Options

```
      --count-total       count total number of records in consensus states to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for consensus-states
      --limit uint        pagination limit of consensus states to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of consensus states to query for
      --page uint         pagination page of consensus states to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of consensus states to query for
      --reverse           results are sorted in descending order
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

- [axelard query ibc client](axelard_query_ibc_client.md)	 - IBC client query subcommands
