## axelard keys

Manage your application's keys

### Synopsis

Keyring management commands. These keys may be in any format supported by the
Tendermint crypto library and can be used by light-clients, full nodes, or any other application
that needs to sign with a private key.

The keyring supports the following backends:

```
os          Uses the operating system's default credentials store.
file        Uses encrypted file-based keystore within the app's configuration directory.
            This keyring will request a password each time it is accessed, which may occur
            multiple times in a single command resulting in repeated password prompts.
kwallet     Uses KDE Wallet Manager as a credentials management application.
pass        Uses the pass command line utility to store and retrieve keys.
test        Stores keys insecurely to disk. It does not prompt for a password to be unlocked
            and it should be use only for testing purposes.
```

kwallet and pass backends depend on external tools. Refer to their respective documentation for more
information:
KWallet     https://github.com/KDE/kwallet
pass        https://www.passwordstore.org/

The pass backend requires GnuPG: https://gnupg.org/

### Options

```
  -h, --help                     help for keys
      --home string              The application home directory (default "$HOME/.axelar")
      --keyring-backend string   Select keyring's backend (os|file|test) (default "file")
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --output string            Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard](axelard.md)	 - Axelar App
- [axelard keys add](axelard_keys_add.md)	 - Add an encrypted private key (either newly generated or recovered), encrypt it, and save to <name> file
- [axelard keys delete](axelard_keys_delete.md)	 - Delete the given keys
- [axelard keys export](axelard_keys_export.md)	 - Export private keys
- [axelard keys import](axelard_keys_import.md)	 - Import private keys into the local keybase
- [axelard keys list](axelard_keys_list.md)	 - List all keys
- [axelard keys migrate](axelard_keys_migrate.md)	 - Migrate keys from the legacy (db-based) Keybase
- [axelard keys mnemonic](axelard_keys_mnemonic.md)	 - Compute the bip39 mnemonic for some input entropy
- [axelard keys parse](axelard_keys_parse.md)	 - Parse address from hex to bech32 and vice versa
- [axelard keys show](axelard_keys_show.md)	 - Retrieve key information by name or address
