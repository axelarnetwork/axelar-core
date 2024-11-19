## axelard tx wasm grant

Grant authorization to an address

### Synopsis

Grant authorization to an address.
Examples:
$ axelard tx grant \<grantee_addr> execution \<contract_addr> --allow-all-messages --max-calls 1 --no-token-transfer --expiration 1667979596

$ axelard tx grant \<grantee_addr> execution \<contract_addr> --allow-all-messages --max-funds 100000uwasm --expiration 1667979596

$ axelard tx grant \<grantee_addr> execution \<contract_addr> --allow-all-messages --max-calls 5 --max-funds 100000uwasm --expiration 1667979596

```
axelard tx wasm grant [grantee] [message_type="execution"|"migration"] [contract_addr_bech32] --allow-raw-msgs [msg1,msg2,...] --allow-msg-keys [key1,key2,...] --allow-all-messages [flags]
```

### Options

```
  -a, --account-number uint      The account number of the signing account (offline mode only)
      --allow-all-messages       Allow all messages
      --allow-msg-keys strings   Allowed msg keys
      --allow-raw-msgs strings   Allowed raw msgs
  -b, --broadcast-mode string    Transaction broadcasting mode (sync|async|block) (default "block")
      --dry-run                  ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)
      --expiration int           The Unix timestamp.
      --fee-account string       Fee account pays fees for the transaction instead of deducting from the signer
      --fees string              Fees to pay along with transaction; eg: 10uatom
      --from string              Name or address of private key with which to sign
      --gas string               gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically (default 200000)
      --gas-adjustment float     adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
      --gas-prices string        Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom) (default "0.007uaxl")
      --generate-only            Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible)
  -h, --help                     help for grant
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test|memory) (default "file")
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                   Use a connected Ledger device
      --max-calls uint           Maximal number of calls to the contract
      --max-funds string         Maximal amount of tokens transferable to the contract.
      --no-token-transfer        Don't allow token transfer
      --node string              <host>:<port> to tendermint rpc interface for this chain (default "tcp://localhost:26657")
      --note string              Note to add a description to the transaction (previously --memo)
      --offline                  Offline mode (does not allow any online functionality
  -o, --output string            Output format (text|json) (default "json")
  -s, --sequence uint            The sequence number of the signing account (offline mode only)
      --sign-mode string         Choose sign mode (direct|amino-json), this is an advanced feature
      --timeout-height uint      Set a block timeout height to prevent the tx from being committed past a certain height
  -y, --yes                      Skip tx broadcasting prompt confirmation (default true)
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID (default "axelar")
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

- [axelard tx wasm](axelard_tx_wasm.md) - Wasm transaction subcommands
