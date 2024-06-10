## axelard tx ibc client misbehaviour

submit a client misbehaviour

### Synopsis

submit a client misbehaviour to prevent future updates

```
axelard tx ibc client misbehaviour [path/to/misbehaviour.json] [flags]
```

### Examples

```
<appd> tx ibc client misbehaviour [path/to/misbehaviour.json] --from node0 --home ../node0/<app>cli --chain-id $CID
```

### Options

```
  -h, --help   help for misbehaviour
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

- [axelard tx ibc client](axelard_tx_ibc_client.md)	 - IBC client transaction subcommands
