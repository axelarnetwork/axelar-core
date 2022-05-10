# Genesis sync

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import MarkdownPath from '../../components/markdown'
import Tabs from '../../components/tabs'
import CodeBlock from '../../components/code-block'

Start your Axelar node and download the blockchain from scratch.

<Callout emoji="💡">
  Tip: Looking for instructions using the old script `node.sh`?  See [here](join-genesis-old).
</Callout>

<Callout emoji="💡">
  Tip: These instructions syncronize your Axelar node from scratch using the Axelar peer-to-peer network. You can syncronize your node more quickly by downloading a recent snapshot of the blockchain as per [Quick sync](join).
</Callout>

## Prerequisites

- [CLI configuration](config-cli).

## Follow the upgrade path

Configure your system as per [Node configuration](config-node) except specify the correct version of `axelard` according to your network and position in the upgrade path:

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

Example: the first step on the upgrade path.

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"./scripts/setup-node.sh -n mainnet -a v0.10.7"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"./scripts/setup-node.sh -n testnet -a v0.13.6"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"./scripts/setup-node.sh -n testnet-2 -a v0.17.3"}
</CodeBlock>
}
]} />

Start your node with the newly configured `axelard` version:

```bash
axelard start --home $AXELARD_HOME
```

Your Axelar node will resume downloading the blockchain.

After your blockchain has reached `UPGRADE_HEIGHT` you will see a panic in the logs like

```
panic: UPGRADE {NAME} NEEDED at height: {UPGRADE_HEIGHT}:
```

Repeat this process for each entry in the upgrade path.
