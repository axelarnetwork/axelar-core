## axelarcli keys parse

Parse address from hex to bech32 and vice versa

### Synopsis

Convert and print to stdout key addresses and fingerprints from
hexadecimal into bech32 cosmos prefixed format and vice versa.

```
axelarcli keys parse <hex-or-bech32-address> [flags]
```

### Options

```
  -h, --help     help for parse
      --indent   Indent JSON output
```

### Options inherited from parent commands

```
      --chain-id string          Network ID of tendermint node
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
```

### SEE ALSO

- [axelarcli keys](axelarcli_keys.md)	 - Add or view local private keys
