## axelard query wasm contract-state

Querying commands for the wasm module

```
axelard query wasm contract-state [flags]
```

### Options

```
  -h, --help   help for contract-state
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

- [axelard query wasm](axelard_query_wasm.md) - Querying commands for the wasm module
- [axelard query wasm contract-state all](axelard_query_wasm_contract-state_all.md) - Prints out all internal state of a contract given its address
- [axelard query wasm contract-state raw](axelard_query_wasm_contract-state_raw.md) - Prints out internal state for key of a contract given its address
- [axelard query wasm contract-state smart](axelard_query_wasm_contract-state_smart.md) - Calls contract with given address with query data and prints the returned result
