# Stake AXL tokens

Stake AXL tokens on the Axelar network.

Choose an amount `{STAKE_AMOUNT}` of AXL tokens you wish to stake. `{STAKE_AMOUNT}` is denominated in `uaxl` where `1 AXL = 1000000 uaxl`.

- You need at least 1 AXL to participate in consensus on the Axelar network
- You need enough stake to get into the "active set" of size 50: if 50 or more other validators have more stake than you then you cannot participate in consensus.
- Optional: you need at least 2% of total bonded stake to participate in multi-party cryptography protocols with other validators.

Choose a moniker `{MY_MONIKER}` for your validator. There are many other parameters you may choose for your validator. For simplicity these instructions specify default values for all other parameters.

Make your `validator` account into an Axelar validator by staking AXL tokens:

```bash
$AXELARD_HOME/bin/axelard tx staking create-validator --amount {STAKE_AMOUNT}uaxl --moniker "{MY_MONIKER}" --commission-rate="0.10" --commission-max-rate="0.20" --commission-max-change-rate="0.01" --min-self-delegation="1" --pubkey="$(axelard tendermint show-validator)" --from validator
```

## Optional: Learn your valoper address

Learn the `{VALOPER_ADDR}` address associated with your `validator` account

```bash
$AXELARD_HOME/bin/axelard keys show validator -a --bech val --home $AXELARD_HOME
```

## Optional: check the stake amount delegated to your validator

```bash
$AXELARD_HOME/bin/axelard q staking validator {VALOPER_ADDR} | grep tokens
```

## Optional: delegate additional stake to your validator

```bash
$AXELARD_HOME/bin/axelard tx staking delegate {VALOPER_ADDR} {STAKE_AMOUNT}uaxl --from validator
```
