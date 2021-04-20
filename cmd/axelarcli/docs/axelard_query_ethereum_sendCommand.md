## axelard query ethereum sendCommand

Send a transaction signed by \[fromAddress\] that executes the command \[commandID\] to Axelar Gateway

```
axelard query ethereum sendCommand [commandID] [fromAddress] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for sendCommand
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID (default "axelar")
```

### SEE ALSO

- [axelard query ethereum](axelard_query_ethereum.md)	 - Querying commands for the ethereum module
