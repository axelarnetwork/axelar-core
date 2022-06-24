import Callout from 'nextra-theme-docs/callout'

# Rotation of mnemonics in tofnd

Starting from `axelard` `v0.17.3+` and `tofnd` `v0.10.1+`, validators can generate a new `tofnd` mnemonic
to slowly rotate out their old `tofnd` mnemonics for improved security.
New Axelar key rotations will automatically use the most recent mnemonic generated.

<Callout type="warning" emoji="⚠️">
  Caution: A validator needs to make sure their old `tofnd` mnemonics are still backed up.
  These mnemonics are still in use until the keys generated from them are considered "old" by the Axelar network.
</Callout>

A key becomes _old_ after `x` subsequent key rotations for that EVM chain. (Currently `x=7`.)

```bash
# Kill vald/tofnd processes
pkill -9 -f "vald"
pkill -f "tofnd"

# Rotate tofnd mnemonic, the new mnemonic is exported automatically
tofnd -m rotate -d $TOFND_HOME

# NOTE: Keep the old mnemonic backups around

# BACKUP the new exported mnemonic and then DELETE the local copy
cp $TOFND_HOME/export ...
rm $TOFND_HOME/export

# Restart vald/tofnd processes as usual
```

After performing the rotation, monitor your validator to make sure it's
still posting heartbeats and there are no unexpected errors in `vald`/`tofnd` logs.
It's also useful to perform a health check.

## Rotation Frequency

Validators are recommended to have processes in place to rotate their `tofnd` mnemonic once every 2 months.

## Recovery of mnmenonics

As before, you can import a `tofnd` mnemonic with `tofnd -m import -d $TOFND_HOME`.
If there are no other mnemonics yet in `tofnd` storage then the imported mnemonic will be treated as the *latest mnemonic*, 
and automatically used for future key ids that are rotated to and any previous key ids it was already a part of.
Each subsequent imported mnemonic is considered as "old" and so only used for any past key ids that corresponded to it.

```bash
# Recover tofnd mnemonics on a fresh state

# Make sure there is no previous tofnd state
rm -r $TOFND_HOME

# Import your latest tofnd mnemonic first
tofnd -m import -d $TOFND_HOME

# Import your remaining old tofnd mnemonics
tofnd -m import -d $TOFND_HOME
```
