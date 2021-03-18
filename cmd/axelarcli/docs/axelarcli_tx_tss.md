## axelarcli tx tss

tss transactions subcommands

```
axelarcli tx tss [flags]
```

### Options

```
  -h, --help   help for tss
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli tx](axelarcli_tx.md)	 - Transactions subcommands
- [axelarcli tx tss deregister](axelarcli_tx_tss_deregister.md)	 - Deregister from participating in any future key generation
- [axelarcli tx tss mk-assign-next](axelarcli_tx_tss_mk-assign-next.md)	 - Assigns a previously created key with \[keyID\] as the next master key for \[chain\]
- [axelarcli tx tss mk-rotate](axelarcli_tx_tss_mk-rotate.md)	 - Rotate the given chain from the old master key to the previously created one (see mk-refresh)
- [axelarcli tx tss start-keygen](axelarcli_tx_tss_start-keygen.md)	 - Initiate threshold key generation protocol
