# Launch companion processes

Launch validator companion processes `tofnd` and `vald`.

## Launch tofnd

You may wish to redirect log output to a file. Enter the `tofnd` password on stdin after the process is launched.

```bash
$AXELARD_HOME/bin/tofnd -m existing -d $AXELARD_HOME/tofnd >> $AXELARD_HOME/logs/tofnd.log 2>&1
```

View your logs in real time:

```bash
tail -f $AXELARD_HOME/logs/tofnd.log
```

## Launch vald

Learn the `valoper` address associated with your `validator` account:

```bash
$AXELARD_HOME/bin/axelard keys show validator --bech val -a --home $AXELARD_HOME
```

Let `{VALOPER_ADDR}` denote this address.

Launch `vald`. Here are two ways to do it:

1. **Basic.**

```bash
$AXELARD_HOME/bin/axelard vald-start --validator-addr {VALOPER_ADDR} --chain-id $AXELARD_CHAIN_ID --log_level debug --home $AXELARD_HOME
```

2. **Redirect logs to file.** The password prompt may not be visible because `stdout` is redirected to the log file. In this case you may wish to pipe in your keyring password. See [Keyring backend](../../node/keyring) for more info.

```bash
echo $KEYRING_PASSWORD | $AXELARD_HOME/bin/axelard vald-start --validator-addr {VALOPER_ADDR} --chain-id $AXELARD_CHAIN_ID --log_level debug --home $AXELARD_HOME >> ${AXELARD_HOME}/logs/vald.log 2>&1
```

View your logs in real time (if you selected option 2 above):

```bash
tail -f ${AXELARD_HOME}/logs/vald.log
```