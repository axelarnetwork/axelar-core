## axelarcli keys add

Add an encrypted private key (either newly generated or recovered), encrypt it, and save to disk

### Synopsis

Derive a new private key and encrypt to disk.
Optionally specify a BIP39 mnemonic, a BIP39 passphrase to further secure the mnemonic,
and a bip32 HD path to derive a specific account. The key will be stored under the given name
and encrypted with the given password. The only input that is required is the encryption password.

If run with -i, it will prompt the user for BIP44 path, BIP39 mnemonic, and passphrase.
The flag --recover allows one to recover a key from a seed passphrase.
If run with --dry-run, a key would be generated (or recovered) but not stored to the
local keystore.
Use the --pubkey flag to add arbitrary public keys to the keystore for constructing
multisig transactions.

You can add a multisig key by passing the list of key names you want the public
key to be composed of to the --multisig flag and the minimum number of signatures
required through --multisig-threshold. The keys are sorted by address, unless
the flag --nosort is set.

```
axelarcli keys add <name> [flags]
```

### Options

```
      --account uint32            Account number for HD derivation
      --algo string               Key signing algorithm to generate keys for (default "secp256k1")
      --dry-run                   Perform action, but don't add key to local keystore
      --hd-path string            Manual HD Path derivation (overrides BIP44 config)
  -h, --help                      help for add
      --indent                    Add indent to JSON response
      --index uint32              Address index number for HD derivation
  -i, --interactive               Interactively prompt user for BIP39 passphrase and mnemonic
      --ledger                    Store a local reference to a private key on a Ledger device
      --multisig strings          Construct and store a multisig public key (implies --pubkey)
      --multisig-threshold uint   K out of N required signatures. For use in conjunction with --multisig (default 1)
      --no-backup                 Don't print out seed phrase (if others are watching the terminal)
      --nosort                    Keys passed to --multisig are taken in the order they're supplied
      --pubkey string             Parse a public key in bech32 format and save it to disk
      --recover                   Provide seed phrase to recover existing key instead of creating
```

### Options inherited from parent commands

```
      --chain-id string          Network ID of tendermint node
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
```

### SEE ALSO

- [axelarcli keys](axelarcli_keys.md)	 - Add or view local private keys
