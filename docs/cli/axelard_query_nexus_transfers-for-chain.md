## axelard query nexus transfers-for-chain

Query for account by address

```
axelard query nexus transfers-for-chain [chain] [state (pending|archived|insufficient_amount)] [flags]
```

### Options

```
      --count-total        count total number of records in transfers-for-chain to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for transfers-for-chain
      --limit uint         pagination limit of transfers-for-chain to query for (default 100)
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint        pagination offset of transfers-for-chain to query for
  -o, --output string      Output format (text|json) (default "text")
      --page uint          pagination page of transfers-for-chain to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of transfers-for-chain to query for
      --reverse            results are sorted in descending order
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query nexus](axelard_query_nexus.md)	 - Querying commands for the nexus module

