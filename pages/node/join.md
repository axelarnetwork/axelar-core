# Quick sync (recommended)

import Callout from 'nextra-theme-docs/callout'

Start your Axelar node and download the blockchain.

<Callout emoji="💡">
  Tip: Looking for instructions using the old script `node.sh`?  See [here](join-old).
</Callout>

<Callout emoji="💡">
  Tip: These instructions syncronize your Axelar node quickly by downloading a recent snapshot of the blockchain. If instead you prefer to syncronize your Axelar node using the Axelar peer-to-peer network then see [Genesis sync](join-genesis)
</Callout>

## Prerequisites

- Configure your environment as per [CLI configuration](config-cli) and [Node configuration](config-node).
- Install lz4
- Ensure AXELARD_HOME variable is set in your current session. See https://docs.axelar.dev/node/config-node#home-directory (example AXELARD_HOME="$HOME/.axelar").
- Existing chain state is stored in `${AXELARD_HOME}/data`. Delete this directory if it exists.

## Download the latest Axelar blockchain snapshot

Download the latest Axelar blockchain snapshot for your chosen network (testnet or mainnet) from a provider:

- [quicksync.io](https://quicksync.io/networks/axelar.html)
- [staketab.com](https://services.staketab.com/snapshots/axelar) | [instructions](https://services.staketab.com/snapshots/axelar)

The following instructions assume you downloaded the `default` snapshot from `quicksync.io`.

Let `{SNAPSHOT_FILE}` denote the file name of the snapshot you downloaded. Example file names:

- **Testnet:** `axelartestnet-lisbon-2-default.20220207.2240.tar.lz4`
- **Mainnet:** `axelar-dojo-1-default.20220207.2210.tar.lz4`

Decompress the downloaded snapshot into your `${AXELARD_HOME}/data` directory. (You may need to install `lz4`: [MacOS](https://formulae.brew.sh/formula/lz4) | [Ubuntu](https://snapcraft.io/install/lz4/ubuntu).)

```bash
lz4 -dc --no-sparse {SNAPSHOT_FILE} | tar xfC - $AXELARD_HOME
```

## Start your node

```bash
$AXELARD_HOME/bin/axelard start --home $AXELARD_HOME
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