## axelarcli keys mnemonic

Compute the bip39 mnemonic for some input entropy

### Synopsis

Create a bip39 mnemonic, sometimes called a seed phrase, by reading from the system entropy. To pass your own entropy, use --unsafe-entropy

```
axelarcli keys mnemonic [flags]
```

### Options

```
  -h, --help             help for mnemonic
      --unsafe-entropy   Prompt the user to supply their own entropy, instead of relying on the system
```

### Options inherited from parent commands

```
      --chain-id string          Network ID of tendermint node
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
```

### SEE ALSO

- [axelarcli keys](axelarcli_keys.md)	 - Add or view local private keys
