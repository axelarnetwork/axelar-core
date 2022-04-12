# Register broadcaster proxy

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

Axelar validators exchange messages with one another via the Axelar blockchain. Each validator sends these messages from a separate `broadcaster` account.

<Callout type="warning" emoji="⚠️">
  Caution: A validator can only register one `broadcaster` address throughout its lifetime. This `broadcaster` address cannot be changed after it has been registered. If you need to register a different proxy address then you must also create an entirely new validator.
</Callout>

## Learn your broadcaster account address

Your `broadcaster` address `{BROADCASTER_ADDR}` is stored in a text file:

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock>
      {"~/.axelar/broadcaster.address"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock>
      {"~/.axelar_testnet/broadcaster.address"}
    </CodeBlock>
  }
]} />

## Fund your validator and broadcaster accounts

**Testnet:**
Go to [Axelar faucet](http://faucet.testnet.axelar.dev/) and send some free AXL testnet tokens to both `{BROADCASTER_ADDR}` and `{VALIDATOR_ADDR}`.

## Register your broadcaster account

<Tabs tabs={[
  {
    title: "Mainnet",
    content: <CodeBlock language="bash">
      {"echo my-secret-password | ~/.axelar/bin/axelard tx snapshot register-proxy {BROADCASTER_ADDR} --from validator --chain-id axelar-dojo-1 --home ~/.axelar/.core"}
    </CodeBlock>
  },
  {
    title: "Testnet",
    content: <CodeBlock language="bash">
      {"echo my-secret-password | ~/.axelar_testnet/bin/axelard tx snapshot register-proxy {BROADCASTER_ADDR} --from validator --chain-id axelar-testnet-lisbon-3 --home ~/.axelar_testnet/.core"}
    </CodeBlock>
  }
]} />

## Optional: check your broadcaster registration

```bash
echo my-secret-password | ~/.axelar_testnet/bin/axelard q snapshot proxy $(cat ~/.axelar_testnet/validator.bech)
```
