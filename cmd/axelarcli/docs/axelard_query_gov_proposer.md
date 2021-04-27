## axelard query gov proposer

Query the proposer of a governance proposal

### Synopsis

Query which address proposed a proposal with a given ID.

Example:
$ <appd> query gov proposer 1

```
axelard query gov proposer [proposal-id] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for proposer
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

- [axelard query gov](axelard_query_gov.md)	 - Querying commands for the governance module
