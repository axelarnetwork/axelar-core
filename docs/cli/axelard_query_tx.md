## axelard query tx

Query for a transaction by hash, "<addr>/<seq>" combination or comma-separated signatures in a committed block

### Synopsis

Example:
$ <appd> query tx <hash>
$ <appd> query tx --type=acc_seq <addr>/<sequence>
$ <appd> query tx --type=signature \<sig1_base64>,\<sig2_base64...>

```
axelard query tx --type=[hash|acc_seq|signature] [hash|acc_seq|signature] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for tx
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --type string   The type to be used when querying tx, can be one of "hash", "acc_seq", "signature" (default "hash")
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

- [axelard query](axelard_query.md)	 - Querying subcommands
