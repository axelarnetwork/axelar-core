## axelarcli query ethereum deploy

Receive a raw deploy transaction

### Synopsis

Receive a raw deploy transaction

```
axelarcli query ethereum deploy [smart contract file path] [flags]
```

### Options

```
      --gas-limit uint   default Ethereum gas limit (default 3000000)
      --height int       Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help             help for deploy
      --indent           Add indent to JSON response
      --ledger           Use a connected Ledger device
      --node string      <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node       Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli query ethereum](axelarcli_query_ethereum.md)	 - Querying commands for the ethereum module
