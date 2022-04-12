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

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock language="bash">
     {"cp -r ~/.axelar ~/.axelar_mainnet_backup"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock language="bash">
      {"cp -r ~/.axelar_testnet ~/.axelar_testnet_backup"}
    </CodeBlock>
  }
]} />

## Resume your Axelar node

Resume your stopped Axelar node.

<Callout emoji="💡">
  Tip: If your node is still in `catching_up` mode then you might need to use the `-a` flag in the following command to specify a different version of axelar-core depending on your current progress downloading the blockchain. See [Join the Axelar testnet for the first time](./join).
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
  }
]} />

## Check your AXL balance

Let `{MY_ADDRESS}` denote the address of your `validator` account.

<Callout emoji="💡">
  Tip: Your balance will appear only after you have downloaded the blockchain and exited `catching_up` mode.
</Callout>

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock language="bash">
      {"echo my-secret-password | ~/.axelar/bin/axelard q bank balances {MY_ADDRESS} --home ~/.axelar/.core"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock language="bash">
      {"echo my-secret-password | ~/.axelar_testnet/bin/axelard q bank balances {MY_ADDRESS} --home ~/.axelar_testnet/.core"}
    </CodeBlock>
  }
]} />

If this is a new account then you should see no token balances.

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <Markdown>
      {""}
    </Markdown>
  },
  {
    title: "Testnet",
    content: <Markdown>{`
## Get AXL tokens from the faucet

Get free AXL testnet tokens sent to {MY_ADDRESS} from the [Axelar Testnet Faucet](https://faucet.testnet.axelar.dev/).

Check your balance again to see the tokens you received from the faucet.
    `}</Markdown>
  }
]} />

## Recover your secret keys

Join the network as per [Quick sync](./join), except use the flags `-t path_to_tendermint_key -m path_to_validator_mnemonic -r` (`-r` is to reset the chain). These flags work only on a completely fresh state.
