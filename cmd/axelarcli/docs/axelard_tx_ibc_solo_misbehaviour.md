## axelard tx ibc solo misbehaviour

submit a client misbehaviour

### Synopsis

submit a client misbehaviour to prevent future updates

```
axelard tx ibc solo misbehaviour [path/to/misbehaviour.json] [flags]
```

### Examples

```
<appd> tx ibc solo machine misbehaviour [path/to/misbehaviour.json] --from node0 --home ../node0/<app>cli --chain-id $CID
```

### Options

```
  -h, --help   help for misbehaviour
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard tx ibc solo](axelard_tx_ibc_solo.md)	 - Solo Machine transaction subcommands
