## axelard debug raw-bytes

Convert raw bytes output (eg. \[10 21 13 255\]) to hex

### Synopsis

Convert raw-bytes to hex.

Example:
$ <appd> debug raw-bytes \[72 101 108 108 111 44 32 112 108 97 121 103 114 111 117 110 100\]

```
axelard debug raw-bytes [raw-bytes] [flags]
```

### Options

```
  -h, --help   help for raw-bytes
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
