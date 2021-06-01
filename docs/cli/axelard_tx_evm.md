## axelard tx evm

evm transactions subcommands

```
axelard tx evm [flags]
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

- [axelard tx](axelard_tx.md)	 - Transactions subcommands
- [axelard tx evm add-chain](axelard_tx_evm_add-chain.md)	 - Add a new EVM chain
- [axelard tx evm confirm-erc20-deposit](axelard_tx_evm_confirm-erc20-deposit.md)	 - Confirm an ERC20 deposit in an EVM chain transaction that sent given amount of token to a burner address
- [axelard tx evm confirm-erc20-token](axelard_tx_evm_confirm-erc20-token.md)	 - Confirm an ERC20 token deployment in an EVM chain transaction for a given symbol of token and gateway address
- [axelard tx evm link](axelard_tx_evm_link.md)	 - Link a cross chain address to an EVM chain address created by Axelar
- [axelard tx evm sign](axelard_tx_evm_sign.md)	 - sign a raw EVM chain transaction
- [axelard tx evm sign-burn-tokens](axelard_tx_evm_sign-burn-tokens.md)	 - Sign burn command for all confirmed token deposits in an EVM chain
- [axelard tx evm sign-deploy-token](axelard_tx_evm_sign-deploy-token.md)	 - Signs the call data to deploy a token with the AxelarGateway contract
- [axelard tx evm sign-pending-transfers](axelard_tx_evm_sign-pending-transfers.md)	 - Sign all pending transfers to an EVM chain
- [axelard tx evm transfer-ownership](axelard_tx_evm_transfer-ownership.md)	 - Sign transfer ownership command for an EVM chain contract
