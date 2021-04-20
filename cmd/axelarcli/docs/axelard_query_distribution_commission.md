## axelard query distribution commission

Query distribution validator commission

### Synopsis

Query validator commission rewards from delegators to that validator.

Example:
$ <appd> query distribution commission axelarvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
axelard query distribution commission [validator] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for commission
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query distribution](axelard_query_distribution.md)	 - Querying commands for the distribution module
