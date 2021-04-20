## axelard init

Initialize private validator, p2p, genesis, and application configuration files

### Synopsis

Initialize validators's and node's configuration files.

```
axelard init [moniker] [flags]
```

### Options

```
      --chain-id string   genesis file chain-id, if left blank will be randomly created (default "axelar")
  -h, --help              help for init
      --home string       node's home directory (default "/Users/chris/.axelar")
  -o, --overwrite         overwrite the genesis.json file
      --recover           provide seed phrase to recover existing key instead of creating
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
