# Old instructions: Quick sync (recommended)

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../components/tabs'
import CodeBlock from '../../components/code-block'

Start your Axelar node and download the blockchain.

<Callout emoji="💡">
  Tip: These instructions syncronize your Axelar node quickly by downloading a recent snapshot of the blockchain. If instead you prefer to syncronize your Axelar node using the Axelar peer-to-peer network then see [Genesis sync](./join-genesis)
</Callout>

## Prerequisites

- **Operating system:** MacOS or Ubuntu (tested on 18.04)
- **Hardware:** 4 cores, 8-16GB RAM, 512 GB drive, arm64 or amd64. Recommended 6-8 cores, 16-32 GB RAM, 1 TB+ drive.
- **Software:**
  - Install [`jq`](https://stedolan.github.io/jq/download/).
  - Install `lz4`: [MacOS](https://formulae.brew.sh/formula/lz4) | [Ubuntu](https://snapcraft.io/install/lz4/ubuntu)
  - Increase the maximum number of open files on your system. Example: `ulimit -n 16384`. You may wish to add this command to your shell profile so that you don't need to execute it next time you restart your machine.
- You have configured your environment for `axelard` CLI commands as per [Configure your environment](config).

## Choose a keyring password

Your Axelar keyring is encrypted with a password you choose. Learn more about keyring passwords at [Keyring backend](keyring.md).

In what follows you will execute a shell script `node.sh` to join the Axelar network. Your keyring password is supplied to `node.sh` via a `KEYRING_PASSWORD` environment variable. The script `node.sh` enforces an 8-character minimum for your keyring password.

## Join the Axelar network

Clone the [`axelerate-community`](https://github.com/axelarnetwork/axelarate-community) repo:

```bash
git clone https://github.com/axelarnetwork/axelarate-community.git
cd axelarate-community
```

<Tabs tabs={[
{
title: "Mainnet",
content: <div>
Launch a new Axelar mainnet node with version <Markdown>`0.10.7`</Markdown> of axelar-core:
<CodeBlock language="bash">
{"./scripts/node.sh -a v0.10.7 -n mainnet"}
</CodeBlock>
Your Axelar node will initialize your data folder <Markdown>`~/.axelar`</Markdown>

</div>
},
{
title: "Testnet",
content: <div>
Launch a new Axelar testnet node with version <Markdown>`0.13.6`</Markdown> of axelar-core:
<CodeBlock language="bash">
{"./scripts/node.sh -a v0.13.6 -n testnet"}
</CodeBlock>
Your Axelar node will initialize your data folder <Markdown>`~/.axelar_testnet`</Markdown>
</div>
},
{
title: "Testnet-2",
content: <div>
Launch a new Axelar testnet node with version <Markdown>`0.17.0`</Markdown> of axelar-core:
<CodeBlock language="bash">
{"./scripts/node.sh -a v0.17.0 -n testnet-2"}
</CodeBlock>
Your Axelar node will initialize your data folder <Markdown>`~/.axelar_testnet-2`</Markdown>
</div>
}
]} />

To recover your secret keys from mnemonics, use `-t path_to_tendermint_key -m path_to_validator_mnemonic -r` (`-r` is to reset the chain). These flags work only on a completely fresh state.

Then your Axelar node will begin downloading blocks in the blockchain one-by-one.

## Backup your secret keys

BACKUP and DELETE the `validator` account secret mnemonic:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock>
{"~/.axelar/validator.txt"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock>
{"~/.axelar_testnet/validator.txt"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock>
{"~/.axelar_testnet-2/validator.txt"}
</CodeBlock>
}
]} />

BACKUP but do NOT DELETE the Tendermint consensus secret key (this is needed on node restarts):

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock>
{"~/.axelar/.core/config/priv_validator_key.json"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock>
{"~/.axelar_testnet/.core/config/priv_validator_key.json"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock>
{"~/.axelar_testnet-2/.core/config/priv_validator_key.json"}
</CodeBlock>
}
]} />

## View logs

View the streaming logs for your Axelar node:

In a new terminal window:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"tail -f ~/.axelar/logs/axelard.log"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"tail -f ~/.axelar_testnet/logs/axelard.log"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"tail -f ~/.axelar_testnet-2/logs/axelard.log"}
</CodeBlock>
}
]} />

You should see log messages for each block in the blockchain that your node downloads.

## Stop your node, delete your blockchain data

You will not download the entire blockchain in this way. Instead you will stop your node and swap in a recent snapshot of the entire blockchain.

Stop your currently running Axelar node:

```bash
kill -9 $(pgrep -f "axelard start")
```

Delete your `data` directory:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"rm -r ~/.axelar/.core/data"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"rm -r ~/.axelar_testnet/.core/data"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"rm -r ~/.axelar_testnet-2/.core/data"}
</CodeBlock>
}
]} />

# Download the latest Axelar blockchain snapshot

Download the latest Axelar blockchain snapshot for your chosen network (testnet or mainnet) from a provider:

- [quicksync.io](https://quicksync.io/networks/axelar.html)
- [staketab.com](https://cosmos-snap.staketab.com/axelar/) | [instructions](https://github.com/staketab/nginx-cosmos-snap/blob/main/docs/axelar.md)

The following instructions assume you downloaded the `default` snapshot from `quicksync.io`.

Let `{SNAPSHOT_FILE}` denote the file name of the snapshot you downloaded. Example file names:

- **Testnet:** `axelartestnet-lisbon-2-default.20220207.2240.tar.lz4`
- **Mainnet:** `axelar-dojo-1-default.20220207.2210.tar.lz4`

Decompress the downloaded snapshot into your `data` directory:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"lz4 -dc --no-sparse {SNAPSHOT_FILE} | tar xfC - ~/.axelar/.core"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"lz4 -dc --no-sparse {SNAPSHOT_FILE} | tar xfC - ~/.axelar_testnet/.core"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"lz4 -dc --no-sparse {SNAPSHOT_FILE} | tar xfC - ~/.axelar_testnet-2/.core"}
</CodeBlock>
}
]} />

## Resume your node

Resume your Axelar node with the latest version of axelar-core:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"./scripts/node.sh -n mainnet"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"./scripts/node.sh -n testnet"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"./scripts/node.sh -n testnet-2"}
</CodeBlock>
}
]} />

Your Axelar node will launch and resume downloading the blockchain. You should see log messages for new blocks.

## Test whether your blockchain is downloaded

Eventually your Axelar node will download the entire Axelar blockchain and exit `catching_up` mode. At that time your logs will show a new block added to the blockchain every 5 seconds.

You can test whether your Axelar node has exited `catching_up` mode:

```bash
curl localhost:26657/status | jq '.result.sync_info'
```

Look for the field `catching_up`:

- `true`: you are still downloading the blockchain.
- `false`: you have finished downloading the blockchain.

## Next steps

Congratulations! You joined the Axelar network and downloaded the blockchain.

Learn what you can do with Axelar:

- [Basic node management](./basic)
