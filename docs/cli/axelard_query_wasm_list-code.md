## axelard query wasm list-code

List all wasm bytecode on the chain

### Synopsis

List all wasm bytecode on the chain

```
axelard query wasm list-code [flags]
```

### Options

```
      --count-total       count total number of records in list codes to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for list-code
      --limit uint        pagination limit of list codes to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of list codes to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of list codes to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of list codes to query for
      --reverse           results are sorted in descending order
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

- [axelard query wasm](axelard_query_wasm.md) - Querying commands for the wasm module
