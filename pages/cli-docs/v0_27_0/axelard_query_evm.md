# axelard query evm

Querying commands for the evm module

```
axelard query evm [flags]
```

## Options

```
  -h, --help   help for evm
```

## Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

## SEE ALSO

- [axelard query](/cli-docs/v0_27_0/axelard_query) - Querying subcommands
- [axelard query evm address](/cli-docs/v0_27_0/axelard_query_evm_address) - Returns the EVM address
- [axelard query evm batched-commands](/cli-docs/v0_27_0/axelard_query_evm_batched-commands) - Get the signed batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm burner-info](/cli-docs/v0_27_0/axelard_query_evm_burner-info) - Get information about a burner address
- [axelard query evm bytecode](/cli-docs/v0_27_0/axelard_query_evm_bytecode) - Fetch the bytecode of an EVM contract \[contract\] for chain \[chain\]
- [axelard query evm chains](/cli-docs/v0_27_0/axelard_query_evm_chains) - Return the supported EVM chains by status
- [axelard query evm command](/cli-docs/v0_27_0/axelard_query_evm_command) - Get information about an EVM gateway command given a chain and the command ID
- [axelard query evm confirmation-height](/cli-docs/v0_27_0/axelard_query_evm_confirmation-height) - Returns the minimum confirmation height for the given chain
- [axelard query evm erc20-tokens](axelard_query_evm_erc20-tokens) - Returns the ERC20 tokens for the given chain
- [axelard query evm event](/cli-docs/v0_27_0/axelard_query_evm_event) - Returns an event for the given chain
- [axelard query evm gateway-address](/cli-docs/v0_27_0/axelard_query_evm_gateway-address) - Query the Axelar Gateway contract address
- [axelard query evm latest-batched-commands](/cli-docs/v0_27_0/axelard_query_evm_latest-batched-commands) - Get the latest batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm pending-commands](/cli-docs/v0_27_0/axelard_query_evm_pending-commands) - Get the list of commands not yet added to a batch
- [axelard query evm token-address](/cli-docs/v0_27_0/axelard_query_evm_token-address) - Query a token address by by either symbol or asset
- [axelard query evm token-info](/cli-docs/v0_27_0/axelard_query_evm_token-info) - Returns the info of token by either symbol, asset, or address
