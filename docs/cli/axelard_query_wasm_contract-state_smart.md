## axelard query wasm contract-state smart

Calls contract with given address with query data and prints the returned result

### Synopsis

Calls contract with given address with query data and prints the returned result

```
axelard query wasm contract-state smart [bech32_address] [query] [flags]
```

### Options

```
      --ascii              ascii encoded query argument
      --b64                base64 encoded query argument
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for smart
      --hex                hex encoded query argument
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard query wasm contract-state](axelard_query_wasm_contract-state.md)	 - Querying commands for the wasm module

