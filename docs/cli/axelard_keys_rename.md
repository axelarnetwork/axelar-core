## axelard keys rename

Rename an existing key

### Synopsis

Rename a key from the Keybase backend.

Note that renaming offline or ledger keys will rename
only the public key references stored locally, i.e.
private keys stored in a ledger device cannot be renamed with the CLI.

```
axelard keys rename <old_name> <new_name> [flags]
```

### Options

```
  -h, --help   help for rename
  -y, --yes    Skip confirmation prompt when renaming offline or ledger key references (default true)
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

- [axelard keys](axelard_keys.md) - Manage your application's keys
