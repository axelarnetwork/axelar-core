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
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query](axelard_query.md)	 - Querying subcommands
- [axelard query evm address](axelard_query_evm_address.md)	 - Returns the EVM address
- [axelard query evm batched-commands](axelard_query_evm_batched-commands.md)	 - Get the signed batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm bytecode](axelard_query_evm_bytecode.md)	 - Fetch the bytecodes of an EVM contract \[contract\] for chain \[chain\]
- [axelard query evm deposit-address](axelard_query_evm_deposit-address.md)	 - Returns an evm chain deposit address for a recipient address on another blockchain
- [axelard query evm deposit-state](axelard_query_evm_deposit-state.md)	 - Query the state of a deposit transaction
- [axelard query evm gateway-address](axelard_query_evm_gateway-address.md)	 - Query the Axelar Gateway contract address
- [axelard query evm latest-batched-commands](axelard_query_evm_latest-batched-commands.md)	 - Get the latest batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway
- [axelard query evm signed-tx](axelard_query_evm_signed-tx.md)	 - Fetch an EVM transaction \[txID\] that has been signed by the validators for chain \[chain\]
- [axelard query evm token-address](axelard_query_evm_token-address.md)	 - Query a token address by symbol
