# Old instructions: Genesis sync

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import MarkdownPath from '../../components/markdown'
import Tabs from '../../components/tabs'
import CodeBlock from '../../components/code-block'

Start your Axelar node and download the blockchain.

<Callout type="error" emoji="🔥">
  Important: The Axelar network is under active development. Use at your own risk with funds you're comfortable using. See [Terms of use](/terms-of-use).
</Callout>

<Callout emoji="💡">
  Tip: These instructions syncronize your Axelar node using the Axelar peer-to-peer network. You can syncronize your node more quickly by downloading a recent snapshot of the blockchain as per [Quick sync](./join).
</Callout>

## Prerequisites

- **Operating system:** MacOS or Ubuntu (tested on 18.04)
- **Hardware:** 4 cores, 8-16GB RAM, 512 GB drive, arm64 or amd64. Recommended 6-8 cores, 16-32 GB RAM, 1 TB+ drive.
- **Software:**
  - Install [`jq`](https://stedolan.github.io/jq/download/).
  - Increase the maximum number of open files on your system. Example: `ulimit -n 16384`. You may wish to add this command to your shell profile so that you don't need to execute it next time you restart your machine.
- You have configured your environment for `axelard` CLI commands as per [Configure your environment](config).

## Choose a keyring password

Your Axelar keyring is encrypted with a password you choose. Your password must have at least 8 characters.

In what follows you will execute a shell script to join the Axelar network. Your keyring password is supplied to the shell script via a `KEYRING_PASSWORD` environment variable.

<Callout type="warning" emoji="⚠️">
  Caution: In the following instructions you must substitute your chosen keyring password for `my-secret-password`.
</Callout>

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
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a v0.10.7 -n mainnet"}
</CodeBlock>
Your Axelar node will initialize your data folder <Markdown>`~/.axelar`</Markdown>

</div>
},
{
title: "Testnet",
content: <div>
Launch a new Axelar testnet node with version `0.13.6` of axelar-core:
<CodeBlock language="bash">
KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a v0.13.6
</CodeBlock>
Your Axelar node will initialize your data folder `~/.axelar_testnet`
</div>
},
{
title: "Testnet-2",
content: <div>
Launch a new Axelar testnet node with version `0.17.0` of axelar-core:
<CodeBlock language="bash">
KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a v0.17.0
</CodeBlock>
Your Axelar node will initialize your data folder `~/.axelar_testnet-2`
</div>
}
]} />

To recover your secret keys from mnemonics, use `-t path_to_tendermint_key -m path_to_validator_mnemonic -r` (`-r` is to reset the chain). These flags work only on a completely fresh state.

Your Axelar node will launch and begin downloading the blockchain.

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

## Follow the upgrade path

Your Axelar node will download the blockchain until it reaches the first `UPGRADE_HEIGHT` listed below.

<Tabs tabs={[
{
title: "Mainnet",
content: <MarkdownPath src="/md/mainnet/upgrade-path.md" />
},
{
title: "Testnet",
content: <MarkdownPath src="/md/testnet/upgrade-path.md" />
},
{
title: "Testnet-2",
content: <MarkdownPath src="/md/testnet-2/upgrade-path.md" />
}
]} />

After your blockchain has reached `UPGRADE_HEIGHT` you will see a panic in the logs like

```
panic: UPGRADE {NAME} NEEDED at height: {UPGRADE_HEIGHT}:
```

Launch your Axelar node again with the `CORE_VERSION` listed below:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a {CORE_VERSION} -n mainnet"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a {CORE_VERSION} -n testnet"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -a {CORE_VERSION} -n testnet-2"}
</CodeBlock>
}
]} />

Your Axelar node will launch and resume downloading the blockchain.

Repeat this process for each entry in the table.

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
