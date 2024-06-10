## axelard query evm bytecode

Fetch the bytecode of an EVM contract \[contract\] for chain \[chain\]

### Synopsis

Fetch the bytecode of an EVM contract \[contract\] for chain \[chain\]. The value for \[contract\] can be either 'token' or 'burner'.

```
axelard query evm bytecode [chain] [contract] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for bytecode
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

- [axelard query evm](axelard_query_evm.md)	 - Querying commands for the evm module
