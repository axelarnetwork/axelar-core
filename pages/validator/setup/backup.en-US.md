# Back-up your secret data

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

Back-up your validator mnemonics and secret keys.

<Callout type="error" emoji="🔥">
  Important: The Axelar network is under active development. Use at your own risk with funds you're comfortable using. See [Terms of use](/terms-of-use).
</Callout>

You must store backup copies of the following data in a safe place:

1. `validator` account secret mnemonic
2. Tendermint validator secret key
3. `broadcaster` account secret mnemonic
4. `tofnd` secret mnemonic

Items 1 and 2 were created when you completed [Quick sync](../../node/join).

Items 3 and 4 were created when you completed [Launch validator companion processes for the first time](./vald-tofnd).

## Validator account secret mnemonic

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
  }
]} />

## Tendermint validator secret key

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
  }
]} />

## Broadcaster account secret mnemonic

BACKUP and DELETE the `broadcaster` account secret mnemonic:

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock>
      {"~/.axelar/broadcaster.txt"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock>
      {"~/.axelar_testnet/broadcaster.txt"}
    </CodeBlock>
  }
]} />

## Tofnd secret mnemonic

BACKUP and DELETE the `tofnd` secret mnemonic:

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock>
      {"~/.axelar/.tofnd/import"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock>
      {"~/.axelar_testnet/.tofnd/import"}
    </CodeBlock>
  }
]} />