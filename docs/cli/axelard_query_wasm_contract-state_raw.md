## axelard query wasm contract-state raw

Prints out internal state for key of a contract given its address

### Synopsis

Prints out internal state for of a contract given its address

```
axelard query wasm contract-state raw [bech32_address] [key] [flags]
```

### Options

```
      --ascii           ascii encoded key argument
      --b64             base64 encoded key argument
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for raw
      --hex             hex encoded key argument
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

- [axelard query wasm contract-state](axelard_query_wasm_contract-state.md) - Querying commands for the wasm module
