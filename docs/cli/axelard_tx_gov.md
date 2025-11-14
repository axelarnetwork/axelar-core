## axelard tx gov

Governance transactions subcommands

```
axelard tx gov [flags]
```

### Options

```
  -h, --help   help for gov
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

- [axelard tx](axelard_tx.md) - Transactions subcommands
- [axelard tx gov cancel-proposal](axelard_tx_gov_cancel-proposal.md) - Cancel governance proposal before the voting period ends. Must be signed by the proposal creator.
- [axelard tx gov deposit](axelard_tx_gov_deposit.md) - Deposit tokens for an active proposal
- [axelard tx gov draft-proposal](axelard_tx_gov_draft-proposal.md) - Generate a draft proposal json file. The generated proposal json contains only one message (skeleton).
- [axelard tx gov submit-legacy-proposal](axelard_tx_gov_submit-legacy-proposal.md) - Submit a legacy proposal along with an initial deposit
- [axelard tx gov submit-proposal](axelard_tx_gov_submit-proposal.md) - Submit a proposal along with some messages, metadata and deposit
- [axelard tx gov vote](axelard_tx_gov_vote.md) - Vote for an active proposal, options: yes/no/no_with_veto/abstain
- [axelard tx gov weighted-vote](axelard_tx_gov_weighted-vote.md) - Vote for an active proposal, options: yes/no/no_with_veto/abstain
