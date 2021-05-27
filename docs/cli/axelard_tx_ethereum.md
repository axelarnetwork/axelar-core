## axelard tx ethereum

ethereum transactions subcommands

```
axelard tx ethereum [flags]
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

- [axelard tx](axelard_tx.md)	 - Transactions subcommands
- [axelard tx ethereum add-chain](axelard_tx_ethereum_add-chain.md)	 - Add a evm chain
- [axelard tx ethereum confirm-erc20-deposit](axelard_tx_ethereum_confirm-erc20-deposit.md)	 - Confirm an ERC20 deposit in an Ethereum transaction that sent given amount of token to a burner address
- [axelard tx ethereum confirm-erc20-token](axelard_tx_ethereum_confirm-erc20-token.md)	 - Confirm an ERC20 token deployment in an Ethereum transaction for a given symbol of token and gateway address
- [axelard tx ethereum link](axelard_tx_ethereum_link.md)	 - Link a cross chain address to an ethereum address created by Axelar
- [axelard tx ethereum sign](axelard_tx_ethereum_sign.md)	 - sign a raw Ethereum transaction
- [axelard tx ethereum sign-burn-tokens](axelard_tx_ethereum_sign-burn-tokens.md)	 - Sign burn command for all confirmed Ethereum token deposits
- [axelard tx ethereum sign-deploy-token](axelard_tx_ethereum_sign-deploy-token.md)	 - Signs the call data to deploy a token with the AxelarGateway contract
- [axelard tx ethereum sign-pending-transfers](axelard_tx_ethereum_sign-pending-transfers.md)	 - Sign all pending transfers to Ethereum
- [axelard tx ethereum transfer-ownership](axelard_tx_ethereum_transfer-ownership.md)	 - Sign transfer ownership command for Ethereum contract
