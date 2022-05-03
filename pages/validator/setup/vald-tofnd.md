# Launch companion processes

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

Launch validator companion processes for the first time.

Axelar validators need two companion processes called `vald` and `tofnd`.

## Choose a tofnd password

Similar to your Axelar keyring, your `tofnd` storage is encrypted with a password you choose. Your password must have at least 8 characters.

In what follows you will execute a shell script to launch the companion processes. Your keyring and `tofnd` passwords are supplied to the shell script via `KEYRING_PASSWORD` and `TOFND_PASSWORD` environment variables.

<Callout type="warning" emoji="⚠️">
  Caution: In the following instructions you must substitute your chosen keyring and `tofnd` passwords for `my-secret-password` and `my-tofnd-password`.
</Callout>

## Launch companion processes

Launch `vald`, `tofnd` for the first time:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password TOFND_PASSWORD=my-tofnd-password ./scripts/validator-tools-host.sh -n mainnet"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password TOFND_PASSWORD=my-tofnd-password ./scripts/validator-tools-host.sh"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"KEYRING_PASSWORD=my-secret-password TOFND_PASSWORD=my-tofnd-password ./scripts/validator-tools-host.sh -n testnet-2"}
</CodeBlock>
}
]} />

To recover your secret keys from mnemonics, use `-p path_to_broadcaster_mnemonic -z path_to_tofnd_mnemonic`. These flags work only on a completely fresh state.

<Callout type="error" emoji="☠️">
  Danger: You created new secret key material. You must backup this data. Failure to backup this data could result in loss of funds. See [Backup your secret data](./backup) for detailed instructions.
</Callout>

## View logs

View the streaming logs for `vald`, `tofnd`:

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`tail -f ~/.axelar/logs/vald.log
tail -f ~/.axelar/logs/tofnd.log`}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{`tail -f ~/.axelar_testnet/logs/vald.log
tail -f ~/.axelar_testnet/logs/tofnd.log`}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{`tail -f ~/.axelar_testnet-2/logs/vald.log
tail -f ~/.axelar_testnet-2/logs/tofnd.log`}
</CodeBlock>
}
]} />
