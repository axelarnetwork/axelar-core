# Launch companion processes

Launch validator companion processes `tofnd` and `vald`.

## Launch tofnd

You may wish to redirect log output to a file:

```bash
tofnd -m existing -d {AXELARD_HOME}/tofnd >> {AXELARD_HOME}/logs/tofnd.log 2>&1
```

View your logs in real time:

```bash
tail -f {AXELARD_HOME}/logs/tofnd.log
```

## Launch vald

Learn the `valoper` address associated with your `validator` account:

```bash
axelard keys show validator --bech val -a --home {AXELARD_HOME}
```

Let `{VALOPER_ADDR}` denote this address.

Launch `vald`. You may wish to redirect log output to a file:

```bash
axelard vald-start --validator-addr {VALOPER_ADDR} --chain-id {AXELARD_CHAIN_ID} --log_level debug --home {AXELARD_HOME} >> {AXELARD_HOME}/logs/vald.log 2>&1
```

View your logs in real time:

```bash
tail -f {AXELARD_HOME}/logs/vald.log
```
