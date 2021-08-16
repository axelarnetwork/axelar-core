## axelard set-genesis-tss

Set the genesis parameters for the tss module

```
axelard set-genesis-tss [flags]
```

### Options

```
      --ack-window int         A positive integer representing the time period for validators to submit acknowledgments for a keygen/sign in terms of number of blocks
      --bond-fraction string   The % of stake validators have to bond per key share (e.g., "1/200").
      --corruption string      The corruption threshold with which Axelar Core will run the keygen protocol (e.g., "2/3").
  -h, --help                   help for set-genesis-tss
      --keygen string          The minimum % of stake that must be online to authorize generation of a new key in the system (e.g., "9/10").
      --locking-period int     A positive integer representing the locking period for validators in terms of number of blocks
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

- [axelard](axelard.md)	 - Axelar App
