## axelard debug

Tool for helping with debugging your application

```
axelard debug [flags]
```

### Options

```
  -h, --help   help for debug
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

- [axelard](axelard.md) - Axelar App
- [axelard debug addr](axelard_debug_addr.md) - Convert an address between hex and bech32
- [axelard debug codec](axelard_debug_codec.md) - Tool for helping with debugging your application codec
- [axelard debug prefixes](axelard_debug_prefixes.md) - List prefixes used for Human-Readable Part (HRP) in Bech32
- [axelard debug pubkey](axelard_debug_pubkey.md) - Decode a pubkey from proto JSON
- [axelard debug pubkey-raw](axelard_debug_pubkey-raw.md) - Decode a ED25519 or secp256k1 pubkey from hex, base64, or bech32
- [axelard debug raw-bytes](axelard_debug_raw-bytes.md) - Convert raw bytes output (eg. [10 21 13 255]) to hex
