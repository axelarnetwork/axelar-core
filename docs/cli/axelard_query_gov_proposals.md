## axelard query gov proposals

Query proposals with optional filters

```
axelard query gov proposals [flags]
```

### Examples

```
axelard query gov proposals --depositor cosmos1...
axelard query gov proposals --voter cosmos1...
axelard query gov proposals --proposal-status (unspecified | deposit-period | voting-period | passed | rejected | failed)
```

### Options

```
      --depositor account address or key name                                                                        
      --grpc-addr string                                                                                             the gRPC endpoint to use for this chain
      --grpc-insecure                                                                                                allow gRPC over insecure channels, if not the server must use TLS
      --height int                                                                                                   Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                                                                                                         help for proposals
      --no-indent                                                                                                    Do not indent JSON output
      --node string                                                                                                  <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string                                                                                                Output format (text|json) (default "text")
      --page-count-total                                                                                             
      --page-key binary                                                                                              
      --page-limit uint                                                                                              
      --page-offset uint                                                                                             
      --page-reverse                                                                                                 
      --proposal-status ProposalStatus (unspecified | deposit-period | voting-period | passed | rejected | failed)    (default unspecified)
      --voter account address or key name                                                                            
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query gov](axelard_query_gov.md)	 - Querying commands for the gov module

