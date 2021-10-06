## axelard query staking redelegation

Query a redelegation record based on delegator and a source and destination validator address

### Synopsis

Query a redelegation record for an individual delegator between a source and destination validator.

Example:
$ <appd> query staking redelegation axelar1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p axelarvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm axelarvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
axelard query staking redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for redelegation
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
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

- [axelard query staking](axelard_query_staking.md)	 - Querying commands for the staking module
