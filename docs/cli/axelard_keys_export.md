## axelard keys export

Export private keys

### Synopsis

Export a private key from the local keyring in ASCII-armored encrypted format.

When both the --unarmored-hex and --unsafe flags are selected, cryptographic
private key material is exported in an INSECURE fashion that is designed to
allow users to import their keys in hot wallets. This feature is for advanced
users only that are confident about how to handle private keys work and are
FULLY AWARE OF THE RISKS. If you are unsure, you may want to do some research
and export your keys in ASCII-armored encrypted format.

```
axelard keys export <name> [flags]
```

### Options

```
  -h, --help            help for export
      --unarmored-hex   Export unarmored hex privkey. Requires --unsafe.
      --unsafe          Enable unsafe operations. This flag must be switched on along with all unsafe operation-specific options.
```

### Options inherited from parent commands

```
      --home string              The application home directory (default "$HOME/.axelar")
      --keyring-backend string   Select keyring's backend (os|file|test) (default "test")
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --log_format string        The logging format (json|plain) (default "plain")
      --log_level string         The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string            Output format (text|json) (default "text")
      --trace                    print out full stack trace on errors
```

### SEE ALSO

- [axelard keys](axelard_keys.md)	 - Manage your application's keys
