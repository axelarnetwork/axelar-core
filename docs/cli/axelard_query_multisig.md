## axelard query multisig

Querying commands for the multisig module

```
axelard query multisig [flags]
```

### Options

```
  -h, --help   help for multisig
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

- [axelard query](axelard_query.md)	 - Querying subcommands
- [axelard query multisig key](axelard_query_multisig_key.md)	 - Returns the key of the given ID
- [axelard query multisig key-id](axelard_query_multisig_key-id.md)	 - Returns the key ID assigned to a given chain
- [axelard query multisig keygen-session](axelard_query_multisig_keygen-session.md)	 - Returns the keygen session info for the given key ID
- [axelard query multisig next-key-id](axelard_query_multisig_next-key-id.md)	 - Returns the key ID assigned for the next rotation on a given chain and for the given key role
- [axelard query multisig params](axelard_query_multisig_params.md)	 - Returns the params for the multisig module
