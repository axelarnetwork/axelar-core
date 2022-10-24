# Quick sync

import Callout from 'nextra-theme-docs/callout'

Start your Axelar node using a snapshot.

<Callout emoji="💡">
  Tip: Looking for instructions using the old script `node.sh`?  See [here](join-old).
</Callout>

<Callout emoji="💡">
  Tip: These instructions synchronize your Axelar node quickly by downloading a recent snapshot of the blockchain. If instead you prefer to syncronize your Axelar node using the Axelar peer-to-peer network then see [Genesis sync](join-genesis)
</Callout>

## Prerequisites

- Configure your environment as per [CLI configuration](config-cli) and [Node configuration](config-node).
- Ensure `AXELARD_HOME` variable is set in your current session. See [node config](https://docs.axelar.dev/node/config-node#home-directory) (e.g. `AXELARD_HOME="$HOME/.axelar"`).

## Download the latest Axelar blockchain snapshot

Download the latest Axelar blockchain snapshot for your chosen network (testnet or mainnet) from a provider:

- [quicksync.io](https://quicksync.io/networks/axelar.html)
- [staketab.com](https://services.staketab.com/snapshots/axelar)

The following instructions assume you downloaded the `pruned` snapshot from `quicksync.io`.

Let `SNAPSHOT_FILE` denote the file name of the snapshot you downloaded. Example file names:

- **Mainnet:** `axelar-dojo-1-pruned.xyz.tar.lz4`
- **Testnet:** `axelartestnet-lisbon-3-pruned.xyz.tar.lz4`

Install `lz4`: [MacOS](https://formulae.brew.sh/formula/lz4) | [Ubuntu](https://snapcraft.io/install/lz4/ubuntu)

Decompress the downloaded snapshot into your `${AXELARD_HOME}/data` directory.

```bash
# Ensure AXELARD_HOME env var is set or substitute it below

# Remove any existing data if it exists
rm -r ${AXELARD_HOME}/data

lz4 -dc --no-sparse [SNAPSHOT_FILE] | tar xfC - ${AXELARD_HOME}

# Verify that snapshot data was uncompressed
ls $AXERLARD_HOME/data
```

## Start your node

```bash
$AXELARD_HOME/bin/axelard start [moniker] --home $AXELARD_HOME >> $AXELARD_HOME/logs/axelard.log 2>&1 &
```

Your Axelar node will launch and begin downloading the rest of the blockchain after the snapshot.

## View your logs in real time

```bash
tail -f $AXELARD_HOME/logs/axelard.log
```

You should see the height (representing the downloaded blockchain) increasing in the logs.

```
... executed block height=690578 ...
... executed block height=690579 ...
```
