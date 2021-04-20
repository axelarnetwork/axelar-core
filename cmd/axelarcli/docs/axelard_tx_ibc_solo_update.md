## axelard tx ibc solo update

update existing client with a header

### Synopsis

update existing client with a solo machine header

```
axelard tx ibc solo update [client-id] [path/to/header.json] [flags]
```

### Examples

```
<appd> tx ibc solo machine update [client-id] [path/to/header.json] --from node0 --home ../node0/<app>cli --chain-id $CID
```

### Options

```
  -h, --help   help for update
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard tx ibc solo](axelard_tx_ibc_solo.md)	 - Solo Machine transaction subcommands
