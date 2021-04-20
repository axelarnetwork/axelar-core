## axelard query ethereum gateway-address

Query the Axelar Gateway contract address

```
axelard query ethereum gateway-address [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for gateway-address
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query ethereum](axelard_query_ethereum.md)	 - Querying commands for the ethereum module
