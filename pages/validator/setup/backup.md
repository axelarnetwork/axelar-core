# Back-up your secret data

Back-up your validator mnemonics and secret keys.

You must store backup copies of the following data in a safe place:

1. `validator` account secret mnemonic
2. Tendermint validator secret key
3. `broadcaster` account secret mnemonic
4. `tofnd` secret mnemonic

Items 1 and 2 were created when you completed [Quick sync](../../node/join).

Items 3 and 4 were created when you completed [Launch validator companion processes for the first time](./vald-tofnd).

## Validator account secret mnemonic

BACKUP and DELETE the `validator` account secret mnemonic:

```
$AXELARD_HOME/validator.txt
```

## Tendermint validator secret key

BACKUP but do NOT DELETE the Tendermint consensus secret key (this is needed on node restarts):

```
$AXELARD_HOME/config/priv_validator_key.json
```

## Broadcaster account secret mnemonic

BACKUP and DELETE the `broadcaster` account secret mnemonic:

```
$AXELARD_HOME/broadcaster.txt
```

## Tofnd secret mnemonic

BACKUP and DELETE the `tofnd` secret mnemonic:

```
$AXELARD_HOME/.tofnd/import
```
