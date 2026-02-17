## axelard query evm

Querying commands for the evm module

```
axelard query evm [flags]
```

### Options

```
  -h, --help   help for evm
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

- [axelard query](axelard_query.md) - Querying subcommands
- [axelard query evm address](axelard_query_evm_address.md) - Returns the EVM address
- [axelard query evm batched-commands](axelard_query_evm_batched-commands.md) - Get the signed batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm bytecode](axelard_query_evm_bytecode.md) - Fetch the token bytecode for chain [chain]
- [axelard query evm chains](axelard_query_evm_chains.md) - Return the supported EVM chains by status
- [axelard query evm command](axelard_query_evm_command.md) - Get information about an EVM gateway command given a chain and the command ID
- [axelard query evm confirmation-height](axelard_query_evm_confirmation-height.md) - Returns the minimum confirmation height for the given chain
- [axelard query evm erc20-tokens](axelard_query_evm_erc20-tokens.md) - Returns the ERC20 tokens for the given chain
- [axelard query evm event](axelard_query_evm_event.md) - Returns an event for the given chain
- [axelard query evm gateway-address](axelard_query_evm_gateway-address.md) - Query the Axelar Gateway contract address
- [axelard query evm latest-batched-commands](axelard_query_evm_latest-batched-commands.md) - Get the latest batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm params](axelard_query_evm_params.md) - Returns the params for the evm module
- [axelard query evm pending-commands](axelard_query_evm_pending-commands.md) - Get the list of commands not yet added to a batch
- [axelard query evm token-info](axelard_query_evm_token-info.md) - Returns the info of token by either symbol, asset, or address
