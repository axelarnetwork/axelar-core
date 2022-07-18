## axelard query staking delegation

Query a delegation based on address and validator address

### Synopsis

Query delegations for an individual delegator on an individual validator.

Example:
$ <appd> query staking delegation axelar1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p axelarvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
axelard query staking delegation [delegator-addr] [validator-addr] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for delegation
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
