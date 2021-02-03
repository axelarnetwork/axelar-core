## axelarcli keys

Add or view local private keys

### Synopsis

Keys allows you to manage your local keystore for tendermint.

```
These keys may be in any format supported by go-crypto and can be
used by light-clients, full nodes, or any other application that
needs to sign with a private key.
```

### Options

```
  -h, --help                     help for keys
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
```

### Options inherited from parent commands

```
      --chain-id string   Network ID of tendermint node
```

### SEE ALSO

- [axelarcli](axelarcli.md)	 - Axelar Client
- [axelarcli keys add](axelarcli_keys_add.md)	 - Add an encrypted private key (either newly generated or recovered), encrypt it, and save to disk
- [axelarcli keys delete](axelarcli_keys_delete.md)	 - Delete the given keys
- [axelarcli keys export](axelarcli_keys_export.md)	 - Export private keys
- [axelarcli keys import](axelarcli_keys_import.md)	 - Import private keys into the local keybase
- [axelarcli keys list](axelarcli_keys_list.md)	 - List all keys
- [axelarcli keys migrate](axelarcli_keys_migrate.md)	 - Migrate keys from the legacy (db-based) Keybase
- [axelarcli keys mnemonic](axelarcli_keys_mnemonic.md)	 - Compute the bip39 mnemonic for some input entropy
- [axelarcli keys parse](axelarcli_keys_parse.md)	 - Parse address from hex to bech32 and vice versa
- [axelarcli keys show](axelarcli_keys_show.md)	 - Show key info for the given name
