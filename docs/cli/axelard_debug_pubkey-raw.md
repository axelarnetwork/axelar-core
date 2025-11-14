## axelard debug pubkey-raw

Decode a ED25519 or secp256k1 pubkey from hex, base64, or bech32

### Synopsis

Decode a pubkey from hex, base64, or bech32.

```
axelard debug pubkey-raw [pubkey] -t [{ed25519, secp256k1}] [flags]
```

### Examples

```

axelard debug pubkey-raw 8FCA9D6D1F80947FD5E9A05309259746F5F72541121766D5F921339DD061174A
axelard debug pubkey-raw j8qdbR+AlH/V6aBTCSWXRvX3JUESF2bV+SEzndBhF0o=
axelard debug pubkey-raw cosmospub1zcjduepq3l9f6mglsz28l40f5pfsjfvhgm6lwf2pzgtkd40eyyeem5rpza9q47axrz
			
```

### Options

```
  -h, --help          help for pubkey-raw
  -t, --type string   Pubkey type to decode (oneof secp256k1, ed25519) (default "ed25519")
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

- [axelard debug](axelard_debug.md) - Tool for helping with debugging your application
