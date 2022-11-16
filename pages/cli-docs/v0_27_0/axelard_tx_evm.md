# axelard tx evm

evm transactions subcommands

```
axelard tx evm [flags]
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

- [axelard tx](/cli-docs/v0_27_0/axelard_tx) - Transactions subcommands
- [axelard tx evm add-chain](/cli-docs/v0_27_0/axelard_tx_evm_add-chain) - Add a new EVM chain
- [axelard tx evm confirm-erc20-deposit](axelard_tx_evm_confirm-erc20-deposit) - Confirm ERC20 deposits in an EVM chain transaction to a burner address
- [axelard tx evm confirm-erc20-token](axelard_tx_evm_confirm-erc20-token) - Confirm an ERC20 token deployment in an EVM chain transaction for a given asset of some origin chain and gateway address
- [axelard tx evm confirm-gateway-tx](/cli-docs/v0_27_0/axelard_tx_evm_confirm-gateway-tx) - Confirm a gateway transaction in an EVM chain
- [axelard tx evm confirm-transfer-operatorship](/cli-docs/v0_27_0/axelard_tx_evm_confirm-transfer-operatorship) - Confirm a transfer operatorship in an EVM chain transaction
- [axelard tx evm create-burn-tokens](/cli-docs/v0_27_0/axelard_tx_evm_create-burn-tokens) - Create burn commands for all confirmed token deposits in an EVM chain
- [axelard tx evm create-deploy-token](/cli-docs/v0_27_0/axelard_tx_evm_create-deploy-token) - Create a deploy token command with the AxelarGateway contract
- [axelard tx evm create-pending-transfers](/cli-docs/v0_27_0/axelard_tx_evm_create-pending-transfers) - Create commands for handling all pending transfers to an EVM chain
- [axelard tx evm link](/cli-docs/v0_27_0/axelard_tx_evm_link) - Link a cross chain address to an EVM chain address created by Axelar
- [axelard tx evm retry-event](/cli-docs/v0_27_0/axelard_tx_evm_retry-event) - Retry a failed event
- [axelard tx evm set-gateway](/cli-docs/v0_27_0/axelard_tx_evm_set-gateway) - Set the gateway address for the given evm chain
- [axelard tx evm sign-commands](/cli-docs/v0_27_0/axelard_tx_evm_sign-commands) - Sign pending commands for an EVM chain contract
- [axelard tx evm transfer-operatorship](/cli-docs/v0_27_0/axelard_tx_evm_transfer-operatorship) - Create transfer operatorship command for an EVM chain contract
