# Basic node management

import Callout from 'nextra-theme-docs/callout'

Start and stop your node, test whether your blockchain is downloaded. Backup your keys and chain data. Create an account, check your AXL balance, get AXL tokens from the faucet.

## Prerequisites

- Configure your environment as per [CLI configuration](config-cli) and [Node configuration](config-node).
- Install [lz4](https://lz4.github.io/lz4/)
- Ensure AXELARD_HOME variable is set in your current session. See https://docs.axelar.dev/node/config-node#home-directory (example AXELARD_HOME="$HOME/.axelar").


## Start your Axelar node

You may wish to redirect log output to a file (your command will be launched in background):

```bash
$AXELARD_HOME/bin/axelard start --home $AXELARD_HOME >> $AXELARD_HOME/logs/axelard.log 2>&1 &
```

View your logs in real time:

```bash
tail -f $AXELARD_HOME/logs/axelard.log
```

## Test whether your blockchain is downloaded

Eventually your Axelar node will download the entire Axelar blockchain and exit `catching_up` mode. At that time your logs will show a new block added to the blockchain every 5 seconds.

You can test whether your Axelar node has exited `catching_up` mode:

```bash
$AXELARD_HOME/bin/axelard status
```

Look for the field `catching_up`:

- `true`: you are still downloading the blockchain.
- `false`: you have finished downloading the blockchain.


## Stop your Axelar node

Stop your currently running Axelar node:

```bash
pkill -f "axelard start"
```

## Check your Axelar node status

```bash
ps aux | grep "axelard start"
```

No process should be running.

## Backup and restore your node and validator keys

Each time you start your Axelar node `axelard` will look for the following files in `$AXELARD_HOME/config`:

- `node_key.json` : The p2p identity of your node. Back this up if this is a seed node.
- `priv_validator_key.json` : Validator’s Tendermint consensus key. Back this up if this is a validator node.

These files will be created if they do not already exist. You can restore them from a backup simply by placing them in `$AXELARD_HOME/config` before starting your node.

## Backup your chain data

<Callout type="warning" emoji="⚠️">
  Caution: Your node must be stopped in order to properly backup chain data.
</Callout>

Backup your entire node's state simply by copying the `$AXELARD_HOME` directory:

```bash
cp -r $AXELARD_HOME ${AXELARD_HOME}_backup.$(date +"%Y%m%d_%H%M%S")
```

## Create an account

```bash
$AXELARD_HOME/bin/axelard keys add my_account --home $AXELARD_HOME
```

<Callout type="warning" emoji="⚠️">
  Caution: The above command will print a mnemonic to stdout.  This mnemonic allows you to recover the secret key for your new account.  Be sure to store a backup of the mnemonic in a safe place.
</Callout>

## Learn your address

The public address of your account `my_account` was printed to stdout when you created it. You can display the address at any time:

```bash
$AXELARD_HOME/bin/axelard keys show validator -a --home $AXELARD_HOME
```

## Check your AXL balance

Let `{MY_ADDRESS}` denote the address of your `my_account` account.

<Callout emoji="💡">
  Tip: Your balance will appear only after you have downloaded the blockchain and exited `catching_up` mode.
</Callout>

```bash
$AXELARD_HOME/bin/axelard q bank balances {MY_ADDRESS}
```

If this is a new account then you should see no token balances.

## Get AXL tokens from the faucet

**Testnets:**
Go to the Axelar testnet faucet and send some free AXL testnet tokens to `{MY_ADDRESS}`:

- [Testnet-1 Faucet](https://faucet.testnet.axelar.dev/).
- [Testnet-2 Faucet](https://faucet-casablanca.testnet.axelar.dev/)
