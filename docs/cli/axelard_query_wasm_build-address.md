## axelard query wasm build-address

build contract address

```
axelard query wasm build-address [code-hash] [creator-address] [salt-hex-encoded] [json_encoded_init_args (required when set as fixed)] [flags]
```

### Options

```
      --ascii   ascii encoded salt
      --b64     base64 encoded salt
  -h, --help    help for build-address
      --hex     hex encoded salt
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --output string       Output format (text|json) (default "text")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard query wasm](axelard_query_wasm.md) - Querying commands for the wasm module
