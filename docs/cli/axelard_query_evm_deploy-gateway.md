## axelard query evm deploy-gateway

Obtain a raw transaction for the deployment of Axelar Gateway.

```
axelard query evm deploy-gateway [chain] [flags]
```

### Options

```
      --gas-limit uint     EVM gas limit to use in the transaction (default value is 3000000). Set to 0 to estimate gas limit at the node. (default 3000000)
      --gas-price string   EVM gas price to use in the transaction. If flag is omitted (or value set to 0), the gas price will be suggested by the node (default "0")
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for deploy-gateway
      --node string        <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
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

- [axelard query evm](axelard_query_evm.md)	 - Querying commands for the evm module
