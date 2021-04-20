## axelard migrate

Migrate genesis to a specified target version

### Synopsis

Migrate the source genesis into the target version and print to STDOUT.

Example:
$ <appd> migrate v0.36 /path/to/genesis.json --chain-id=cosmoshub-3 --genesis-time=2019-04-22T17:00:00Z

```
axelard migrate [target-version] [genesis-file] [flags]
```

### Options

```
      --chain-id string       override chain_id with this flag (default "axelar")
      --genesis-time string   override genesis_time with this flag
  -h, --help                  help for migrate
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
