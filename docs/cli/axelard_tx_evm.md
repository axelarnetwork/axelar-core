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
- [axelard tx evm confirm-erc20-deposit](axelard_tx_evm_confirm-erc20-deposit.md)	 - Confirm ERC20 deposits in an EVM chain transaction to a burner address
- [axelard tx evm confirm-erc20-token](axelard_tx_evm_confirm-erc20-token.md)	 - Confirm an ERC20 token deployment in an EVM chain transaction for a given asset of some origin chain and gateway address
- [axelard tx evm confirm-gateway-txs](axelard_tx_evm_confirm-gateway-txs.md)	 - Confirm gateway transactions in an EVM chain
- [axelard tx evm confirm-transfer-operatorship](axelard_tx_evm_confirm-transfer-operatorship.md)	 - Confirm a transfer operatorship in an EVM chain transaction
- [axelard tx evm create-burn-tokens](axelard_tx_evm_create-burn-tokens.md)	 - Create burn commands for all confirmed token deposits in an EVM chain
- [axelard tx evm create-deploy-token](axelard_tx_evm_create-deploy-token.md)	 - Create a deploy token command with the AxelarGateway contract
- [axelard tx evm create-pending-transfers](axelard_tx_evm_create-pending-transfers.md)	 - Create commands for handling all pending transfers to an EVM chain
- [axelard tx evm link](axelard_tx_evm_link.md)	 - Link a cross chain address to an EVM chain address created by Axelar
- [axelard tx evm retry-event](axelard_tx_evm_retry-event.md)	 - Retry a failed event
- [axelard tx evm set-gateway](axelard_tx_evm_set-gateway.md)	 - Set the gateway address for the given evm chain
- [axelard tx evm sign-commands](axelard_tx_evm_sign-commands.md)	 - Sign pending commands for an EVM chain contract
- [axelard tx evm transfer-operatorship](axelard_tx_evm_transfer-operatorship.md)	 - Create transfer operatorship command for an EVM chain contract
