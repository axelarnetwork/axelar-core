## axelard debug pubkey

Decode a ED25519 pubkey from hex, base64, or bech32

### Synopsis

Decode a pubkey from hex, base64, or bech32.

Example:
$ <appd> debug pubkey TWFuIGlzIGRpc3Rpbmd1aXNoZWQsIG5vdCBvbmx5IGJ5IGhpcyByZWFzb24sIGJ1dCBieSB0aGlz
$ <appd> debug pubkey cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg

```
axelard debug pubkey [pubkey] [flags]
```

### Options

```
  -h, --help   help for pubkey
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard debug](axelard_debug.md)	 - Tool for helping with debugging your application
