# CLI configuration

By default, many Axelar CLI commands contain boilerplate material that obfuscates the command and adds complexity to documentation. Here we describe documentation conventions and ways to configure your system so as to reduce this boilerplate.

Example: As per [Send AXL to an EVM chain](../learn/cli/axl-to-evm) you can use the CLI to get a deposit address for cross-chain transfer of AXL tokens via the command

```bash
axelard tx axelarnet link {EVM_CHAIN} {EVM_DEST_ADDR} uaxl --from my_account
```

The above command does not work on a system that has not yet been properly configured. On a fresh, unconfigured system you would instead need to type

```bash
echo $KEYRING_PASSWORD | ~/.axelar_testnet/bin/axelard tx axelarnet link {EVM_CHAIN} {EVM_DEST_ADDR} uaxl --from validator --gas auto --gas-adjustment 1.5 --chain-id axelar-testnet-lisbon-3 --home ~/.axelar_testnet
```

Let us explain the changes:

- **Keyring password.** `echo $KEYRING_PASSWORD` : Commands that use `tx` to post a signed transaction to the network require a password to unlock the keyring. For simplicity documentation elides the keyring password. See [Keyring backend](keyring) for more info.
- **Path to `axelard`.** `~/.axelar_testnet/bin/axelard` : You can avoid the need to type the full path to `axelard` by putting `axelard` in your `PATH` environment variable or defining a shell alias.
- **Gas.** `--gas auto --gas-adjustment 1.5` : Commands that use `tx` to post a signed transaction to the network must pay transaction fees called _gas_. By default `axelard` estimates the gas fee but sometimes this estimation is inaccurate and the transaction fails due to insufficient gas fee. These gas flags instruct `axelard` to estimate gas more accurately, then increase the maximum gas for the transaction by a factor you specify (`1.5` in this case) so that your transaction is more likely to have sufficient gas even if actual network fees exceed the estimate. For simplicity documentation elides these flags.
- **Chain ID.** `--chain-id axelar-testnet-lisbon-3` : Commands that use `tx` to post a signed transaction to the network must specify the `chain-id` for the transaction. You could instead set the `AXELARD_CHAIN_ID` environment variable.
- **Home directory.** `--home ~/.axelar_testnet` : Many commands must specify the path to on-disk storage for `axelard`. You could instead set the `AXELARD_HOME` environment variable. The default value is `~/.axelar`.
