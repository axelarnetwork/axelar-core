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
- [axelard query evm bytecode](axelard_query_evm_bytecode.md)	 - Fetch the bytecodes of an EVM contract \[contract\] for chain \[chain\]
- [axelard query evm command](axelard_query_evm_command.md)	 - Get the signed command data that can be wrapped in an EVM transaction to execute the command \[commandID\] on Axelar Gateway
- [axelard query evm deploy-gateway](axelard_query_evm_deploy-gateway.md)	 - Obtain a raw transaction for the deployment of Axelar Gateway.
- [axelard query evm deposit-address](axelard_query_evm_deposit-address.md)	 - Returns an evm chain deposit address for a recipient address on another blockchain
- [axelard query evm gateway-address](axelard_query_evm_gateway-address.md)	 - Query the Axelar Gateway contract address
- [axelard query evm master-address](axelard_query_evm_master-address.md)	 - Returns the EVM address of the current master key, and optionally the key's ID
- [axelard query evm sendCommand](axelard_query_evm_sendCommand.md)	 - Send a transaction signed by \[fromAddress\] that executes the command \[commandID\] to Axelar Gateway
- [axelard query evm sendTx](axelard_query_evm_sendTx.md)	 - Send a transaction that spends tx \[txID\] to chain \[chain\]
- [axelard query evm signedTx](axelard_query_evm_signedTx.md)	 - Fetch an EVM transaction \[txID\] that has been signed by the validators for chain \[chain\]
- [axelard query evm token-address](axelard_query_evm_token-address.md)	 - Query a token address by symbol
