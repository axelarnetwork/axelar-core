## axelard query tss

Querying commands for the tss module

```
axelard query tss [flags]
```

### Options

```
  -h, --help   help for tss
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
- [axelard query tss active-old-keys](axelard_query_tss_active-old-keys.md)	 - Query active old key IDs by validator
- [axelard query tss active-old-keys-by-validator](axelard_query_tss_active-old-keys-by-validator.md)	 - Query active old key IDs by validator
- [axelard query tss assignable-key](axelard_query_tss_assignable-key.md)	 - Returns the true if a key can be assigned for the next rotation on a given chain and for the given key role
- [axelard query tss deactivated-operators](axelard_query_tss_deactivated-operators.md)	 - Fetch the list of deactivated operator addresses
- [axelard query tss external-key-id](axelard_query_tss_external-key-id.md)	 - Returns the key IDs of the current external keys for the given chain
- [axelard query tss key](axelard_query_tss_key.md)	 - Query a key by key ID
- [axelard query tss key-id](axelard_query_tss_key-id.md)	 - Query the keyID using keyChain and keyRole
- [axelard query tss key-shares-by-key-id](axelard_query_tss_key-shares-by-key-id.md)	 - Query key shares information by key ID
- [axelard query tss key-shares-by-validator](axelard_query_tss_key-shares-by-validator.md)	 - Query key shares information by validator
- [axelard query tss next-key-id](axelard_query_tss_next-key-id.md)	 - Returns the key ID assigned for the next rotation on a given chain and for the given key role
- [axelard query tss recover](axelard_query_tss_recover.md)	 - Attempt to recover the shares for the specified key ID
- [axelard query tss signature](axelard_query_tss_signature.md)	 - Query a signature by sig ID
