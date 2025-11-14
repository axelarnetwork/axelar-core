## axelard query blocks

Query for paginated blocks that match a set of events

### Synopsis

Search for blocks that match the exact given events where results are paginated.
The events query is directly passed to CometBFT's RPC BlockSearch method and must
conform to CometBFT's query syntax.
Please refer to each module's documentation for the full set of events to query
for. Each module documents its respective events under 'xx_events.md'.


```
axelard query blocks [flags]
```

### Examples

```
$ axelard query blocks --query "message.sender='cosmos1...' AND block.height > 7" --page 1 --limit 30 --order_by asc
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for blocks
      --limit int          Query number of transactions results per page returned (default 100)
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
      --order_by string    The ordering semantics (asc|dsc)
  -o, --output string      Output format (text|json) (default "text")
      --page int           Query a specific page of paginated results (default 1)
      --query string       The blocks events query per CometBFT's query semantics
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query](axelard_query.md)	 - Querying subcommands

