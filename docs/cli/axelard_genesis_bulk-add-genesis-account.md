## axelard genesis bulk-add-genesis-account

Bulk add genesis accounts to genesis.json

### Synopsis

Add genesis accounts in bulk to genesis.json. The provided account must specify
the account address and a list of initial coins. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.


```
axelard genesis bulk-add-genesis-account [/file/path.json] [flags]
```

### Examples

```
bulk-add-genesis-account accounts.json
where accounts.json is:
[
    {
        "address": "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
        "coins": [
            { "denom": "umuon", "amount": "100000000" },
            { "denom": "stake", "amount": "200000000" }
        ]
    },
    {
        "address": "cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg",
        "coins": [
            { "denom": "umuon", "amount": "500000000" }
        ],
        "vesting_amt": [
            { "denom": "umuon", "amount": "400000000" }
        ],
        "vesting_start": 1724711478,
        "vesting_end": 1914013878
    }
]

```

### Options

```
      --append             append the coins to an account already in the genesis.json file
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for bulk-add-genesis-account
      --home string        The application home directory (default "$HOME/.axelar")
      --node string        <host>:<port> to CometBFT RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string      Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [axelard genesis](axelard_genesis.md)	 - Application's genesis-related subcommands

