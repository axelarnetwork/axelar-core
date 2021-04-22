## axelard query ibc connection end

Query stored connection end

### Synopsis

Query stored connection end

```
axelard query ibc connection end [connection-id] [flags]
```

### Examples

```
<appd> query ibc connection end [connection-id]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for end
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

- [axelard query ibc connection](axelard_query_ibc_connection.md)	 - IBC connection query subcommands
