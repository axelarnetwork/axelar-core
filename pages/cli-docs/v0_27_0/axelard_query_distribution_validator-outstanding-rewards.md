# axelard query distribution validator-outstanding-rewards

Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations

## Synopsis

Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations.

Example:

```
$ <appd> query distribution validator-outstanding-rewards axelarvaloper1lwnks33xwnmfayc64ycprww49n33mtm92ne
```

```
axelard query distribution validator-outstanding-rewards [validator] [flags]
```

## Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for validator-outstanding-rewards
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

- [axelard query distribution](/cli-docs/v0_27_0/axelard_query_distribution) - Querying commands for the distribution module
