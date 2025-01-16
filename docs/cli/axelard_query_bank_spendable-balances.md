## axelard query bank spendable-balances

Query for account spendable balances by address

```
axelard query bank spendable-balances [address] [flags]
```

### Examples

```
$ axelard query bank spendable-balances [address]
```

### Options

```
      --count-total        count total number of records in spendable balances to query for
      --denom string       The specific balance denomination to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for spendable-balances
      --limit uint         pagination limit of spendable balances to query for (default 100)
      --node string        <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint        pagination offset of spendable balances to query for
  -o, --output string      Output format (text|json) (default "text")
      --page uint          pagination page of spendable balances to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of spendable balances to query for
      --reverse            results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query bank](axelard_query_bank.md) - Querying commands for the bank module
