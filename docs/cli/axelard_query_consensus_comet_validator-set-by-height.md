## axelard query consensus comet validator-set-by-height

Query for a validator set by height

```
axelard query consensus comet validator-set-by-height [height] [flags]
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for validator-set-by-height
      --no-indent          Do not indent JSON output
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
      --page-count-total   
      --page-key binary    
      --page-limit uint    
      --page-offset uint   
      --page-reverse       
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query consensus comet](axelard_query_consensus_comet.md)	 - Querying commands for the cosmos.base.tendermint.v1beta1.Service service

