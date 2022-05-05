# Basic node management

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../components/tabs'
import CodeBlock from '../../components/code-block'

Stop your node, backup your chain data, resume your node. Check your AXL balance, get AXL tokens from the faucet.

<Callout type="error" emoji="🔥">
  Important: The Axelar network is under active development. Use at your own risk with funds you're comfortable using. See [Terms of use](/terms-of-use).
</Callout>

## Prerequisites

You have launched your Axelar node as per [Quick sync](./join). Perhaps you have not yet completed downloading the blockchain.

## Stop your Axelar node

Stop your currently running Axelar node:

```bash
kill -9 $(pgrep -f "axelard start")
```

## Backup your chain data

<Callout type="warning" emoji="⚠️">
  Caution: Your node must be stopped in order to properly backup chain data.
</Callout>

```bash
cp -r $AXELARD_HOME ${AXELARD_HOME}_backup
```

## Resume your Axelar node

Resume your stopped Axelar node.

<Callout emoji="💡">
  Tip: If your node is still in `catching_up` mode then you might need to use the `-a` flag in the following command to specify a different version of axelar-core depending on your current progress downloading the blockchain. See [Genesis sync](./join-genesis).
</Callout>

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -n mainnet"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password ./scripts/node.sh -n testnet-2"}
</CodeBlock>
}
]} />

## Learn your address

<Callout emoji="💡">
  Tip: A new account named `validator` was automatically created for you when you joined the Axelar network for the first time. This is just a name---you are not (yet) a validator on the Axelar network.
</Callout>

Learn the address of your `validator` account:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"echo my-secret-password | ~/.axelar/bin/axelard keys show validator -a --home ~/.axelar/.core"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"echo my-secret-password | ~/.axelar_testnet/bin/axelard keys show validator -a --home ~/.axelar_testnet/.core"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"echo my-secret-password | ~/.axelar_testnet-2/bin/axelard keys show validator -a --home ~/.axelar_testnet-2/.core"}
</CodeBlock>
}
]} />

## Check your AXL balance

Let `{MY_ADDRESS}` denote the address of your `validator` account.

<Callout emoji="💡">
  Tip: Your balance will appear only after you have downloaded the blockchain and exited `catching_up` mode.
</Callout>

```bash
axelard q bank balances {MY_ADDRESS}
```

If this is a new account then you should see no token balances.

## Get AXL tokens from the faucet

**Testnets:**
Go to the Axelar testnet faucet and send some free AXL testnet tokens to `{MY_ADDRESS}`:

- [Testnet-1 Faucet](https://faucet.testnet.axelar.dev/).
- [Testnet-2 Faucet](https://faucet-casablanca.testnet.axelar.dev/)

## Recover your secret keys

Join the network as per [Quick sync](./join), except use the flags `-t path_to_tendermint_key -m path_to_validator_mnemonic -r` (`-r` is to reset the chain). These flags work only on a completely fresh state.
