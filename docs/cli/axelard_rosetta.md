## axelard rosetta

spin up a rosetta server

```
axelard rosetta [flags]
```

### Options

```
      --addr string                the address rosetta will bind to (default ":8080")
      --bech32-prefix string       address bech32 prefix (default "cosmos")
      --blockchain string          the blockchain type (default "app")
      --denom-to-suggest string    default denom for fee suggestion (default "uatom")
      --enable-fee-suggestion      enable default fee suggestion
      --gas-to-suggest int         default gas for fee suggestion (default 200000)
      --grpc string                the app gRPC endpoint (default "localhost:9090")
      --grpc-types-server string   the app gRPC Server endpoint for proto messages types and reflection
  -h, --help                       help for rosetta
      --network string             the network name (default "network")
      --offline                    run rosetta only with construction API
      --plugin string              plugin folder name
      --prices-to-suggest string   default prices for fee suggestion (default "1uatom,1stake")
      --retries int                the number of retries that will be done before quitting (default 5)
      --symbol-decimals string     symbol decimals mapping (format: base:symbol:decimal, comma-separated, e.g. "uaxl:AXL:6")
      --tendermint string          the CometBFT rpc endpoint, without tcp:// (default "localhost:26657")
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md) - Axelar App
