## axelarcli query bitcoin rawTx

Get a raw transaction that spends \[amount\] of the outpoint \[voutIdx\] of \[txID\] to <recipient> or the next master key in rotation

### Synopsis

Get a raw transaction that spends \[amount\] of the outpoint \[voutIdx\] of \[txID\] to <recipient> or the next master key in rotation

```
axelarcli query bitcoin rawTx [txID:voutIdx] [amount] [recipient] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for rawTx
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli query bitcoin](axelarcli_query_bitcoin.md)	 - bitcoin query subcommands
