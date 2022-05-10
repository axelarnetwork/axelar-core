# Create and backup accounts

Create and backup your validator mnemonics and secret keys.

You must store backup copies of the following data in a safe place:

1. Tendermint validator secret key
2. `validator` account secret mnemonic
3. `broadcaster` account secret mnemonic
4. `tofnd` secret mnemonic

## Backup your Tendermint validator secret key

As described in [Basic node management](../../node/basic) BACKUP the file `${AXELARD_HOME}/config/priv_validator_key.json`.

## Create and backup accounts

Each validator needs two accounts, which we call `validator` and `broadcaster`. Create those accounts and back them up.

```bash
axelard keys add validator --home $AXELARD_HOME
axelard keys add broadcaster --home $AXELARD_HOME
```

As described in [Basic node management](../../node/basic), BACKUP the secret mnemonics for these accounts that are printed to stdout when you crate them.

## Create and backup tofnd mnemonic

Similar to your [Axelar keyring](../../node/keyring), your `tofnd` storage is encrypted with a password you choose. Your password must have at least 8 characters.

Set `tofnd` password and create `tofnd` mnemonic:

```bash
tofnd -m create -d ${AXELARD_HOME}/tofnd
```

BACKUP and DELETE `${AXELARD_HOME}/tofnd/export`.
