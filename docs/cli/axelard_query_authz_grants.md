## axelard query authz grants

Query grants for a granter-grantee pair and optionally a msg-type-url

### Synopsis

Query authorization grants for a granter-grantee pair. If msg-type-url is set, it will select grants only for that msg type.

```
axelard query authz grants [granter-addr] [grantee-addr] <msg-type-url> [flags]
```

### Examples

```
axelard query authz grants cosmos1skj.. cosmos1skjwj.. /cosmos.bank.v1beta1.MsgSend
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for grants
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

- [axelard query authz](axelard_query_authz.md) - Querying commands for the authz module
