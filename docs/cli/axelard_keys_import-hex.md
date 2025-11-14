## axelard keys import-hex

Import private keys into the local keybase

### Synopsis

Import hex encoded private key into the local keybase.
Supported key-types can be obtained with:
axelard list-key-types

```
axelard keys import-hex <name> <hex> [flags]
```

### Options

```
  -h, --help              help for import-hex
      --key-type string   private key signing algorithm kind (default "secp256k1")
```

### Options inherited from parent commands

```
      --home string              directory for config and data (default "$HOME/.axelar")
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test|memory) (default "file")
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --log_format string        The logging format (json|plain) (default "plain")
      --log_level string         The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string            Output format (text|json) (default "text")
      --trace                    print out full stack trace on errors
```

### SEE ALSO

* [axelard keys](axelard_keys.md)	 - Manage your application's keys

