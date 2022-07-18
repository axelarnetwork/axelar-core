## axelard query ibc client status

Query client status

### Synopsis

Query client activity status. Any client without an 'Active' status is considered inactive

```
axelard query ibc client status [client-id] [flags]
```

### Examples

```
<appd> query ibc client status [client-id]
```

### Options

```
  -h, --help   help for status
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

- [axelard query ibc client](axelard_query_ibc_client.md)	 - IBC client query subcommands
