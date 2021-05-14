## axelard query ethereum

Querying commands for the ethereum module

```
axelard query ethereum [flags]
```

### Options

```
  -h, --help   help for ethereum
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
- [axelard query ethereum command](axelard_query_ethereum_command.md)	 - Get the signed command data that can be wrapped in an Ethereum transaction to execute the command \[commandID\] on Axelar Gateway
- [axelard query ethereum deploy-gateway](axelard_query_ethereum_deploy-gateway.md)	 - Obtain a raw transaction for the deployment of Axelar Gateway.
- [axelard query ethereum gateway-address](axelard_query_ethereum_gateway-address.md)	 - Query the Axelar Gateway contract address
- [axelard query ethereum master-address](axelard_query_ethereum_master-address.md)	 - Query an address by key ID
- [axelard query ethereum sendCommand](axelard_query_ethereum_sendCommand.md)	 - Send a transaction signed by \[fromAddress\] that executes the command \[commandID\] to Axelar Gateway
- [axelard query ethereum sendTx](axelard_query_ethereum_sendTx.md)	 - Send a transaction that spends tx \[txID\] to Ethereum
- [axelard query ethereum token-address](axelard_query_ethereum_token-address.md)	 - Query a token address by symbol
