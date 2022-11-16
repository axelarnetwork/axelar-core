# axelard query ibc-transfer denom-trace

Query the denom trace info from a given trace hash or ibc denom

## Synopsis

Query the denom trace info from a given trace hash or ibc denom

```
axelard query ibc-transfer denom-trace [hash/denom] [flags]
```

## Examples

```
<appd> query ibc-transfer denom-trace 27A6394C3F9FF9C9DCF5DFFADF9BB5FE9A37C7E92B006199894CF1824DF9AC7C
```

## Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for denom-trace
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
```

## Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

## SEE ALSO

- [axelard query ibc-transfer](/cli-docs/v0_27_0/axelard_query_ibc-transfer) - IBC fungible token transfer query subcommands
