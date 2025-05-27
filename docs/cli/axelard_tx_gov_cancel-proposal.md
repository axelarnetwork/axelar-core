<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
## axelard tx gov cancel-proposal
========
## axelard tx gov submit-legacy-proposal update-client
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md

Cancel governance proposal before the voting period ends. Must be signed by the proposal creator.

```
<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
axelard tx gov cancel-proposal [proposal-id] [flags]
```

### Examples

```
$ axelard tx gov cancel-proposal 1 --from mykey
========
axelard tx gov submit-legacy-proposal update-client [subject-client-id] [substitute-client-id] [flags]
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md
```

### Options

```
  -a, --account-number uint      The account number of the signing account (offline mode only)
      --aux                      Generate aux signer data instead of sending a tx
  -b, --broadcast-mode string    Transaction broadcasting mode (sync|async) (default "sync")
<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
      --chain-id string          The network chain ID
========
      --chain-id string          The network chain ID (default "axelar")
      --deposit string           deposit of proposal
      --description string       description of proposal
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md
      --dry-run                  ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it (when enabled, the local Keybase is not accessible)
      --fee-granter string       Fee granter grants fees for the transaction
      --fee-payer string         Fee payer pays fees for the transaction instead of deducting from the signer
      --fees string              Fees to pay along with transaction; eg: 10uatom
      --from string              Name or address of private key with which to sign
      --gas string               gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically. Note: "auto" option doesn't always report accurate results. Set a valid coin value to adjust the result. Can be used instead of "fees". (default 200000)
      --gas-adjustment float     adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
      --gas-prices string        Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)
      --generate-only            Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase only accessed when providing a key name)
  -h, --help                     help for cancel-proposal
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test|memory) (default "os")
========
      --gas-prices string        Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom) (default "0.007uaxl")
      --generate-only            Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase only accessed when providing a key name)
  -h, --help                     help for update-client
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test|memory) (default "file")
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                   Use a connected Ledger device
      --node string              <host>:<port> to CometBFT rpc interface for this chain (default "tcp://localhost:26657")
      --note string              Note to add a description to the transaction (previously --memo)
      --offline                  Offline mode (does not allow any online functionality)
  -o, --output string            Output format (text|json) (default "json")
  -s, --sequence uint            The sequence number of the signing account (offline mode only)
<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
      --sign-mode string         Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature
      --timeout-height uint      Set a block timeout height to prevent the tx from being committed past a certain height
      --tip string               Tip is the amount that is going to be transferred to the fee payer on the target chain. This flag is only valid when used with --aux, and is ignored if the target chain didn't enable the TipDecorator
  -y, --yes                      Skip tx broadcasting prompt confirmation
========
      --sign-mode string         Choose sign mode (direct|amino-json|direct-aux), this is an advanced feature
      --timeout-height uint      Set a block timeout height to prevent the tx from being committed past a certain height
      --tip string               Tip is the amount that is going to be transferred to the fee payer on the target chain. This flag is only valid when used with --aux, and is ignored if the target chain didn't enable the TipDecorator
      --title string             title of proposal
  -y, --yes                      Skip tx broadcasting prompt confirmation (default true)
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md
```

### Options inherited from parent commands

```
      --home string         directory for config and data (default "$HOME/.axelar")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors
```

### SEE ALSO

<<<<<<<< HEAD:docs/cli/axelard_tx_gov_cancel-proposal.md
- [axelard tx gov](axelard_tx_gov.md) - Governance transactions subcommands
========
- [axelard tx gov submit-legacy-proposal](axelard_tx_gov_submit-legacy-proposal.md) - Submit a legacy proposal along with an initial deposit
>>>>>>>> cosmos-sdk-v0.47:docs/cli/axelard_tx_gov_submit-legacy-proposal_update-client.md
