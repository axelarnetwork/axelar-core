# Keyring backend

import Callout from 'nextra-theme-docs/callout'

Many Axeler CLI commands require an Axelar account controlled by a secret key. Your secret key must be stored securely so as to minimize the risk of exposure to an attacker.

Like every Cosmos-based network, Axelar nodes store secret keys in a _keyring_. The keyring can be configured with one of several _backend_ implementations. Learn more about keyring backend configuration from the [Cosmos keyring documentation](https://docs.cosmos.network/v0.44/run-node/keyring.html).

Axelar nodes use the `file` keyring backend by default. This means that your secret keys are stored in a password-encrypted file on disk. Under the `file` backend, you must provide your keyring password each time you execute certain Axelar CLI commands.

<Callout type="warning" emoji="⚠️">
  Protect your keyring password: There are several methods to provide your password for Axelar CLI commands. Each method comes with its own security and convenience properties. Whichever method you choose, be sure to follow best practices to keep your keyring password safe.
</Callout>

## Prerequisites

- Configure your environment as per [CLI configuration](config-cli) and [Node configuration](config-node).
- Ensure AXELARD_HOME variable is set in your current session. See https://docs.axelar.dev/node/config-node#home-directory (example AXELARD_HOME="$HOME/.axelar").

## Manual password entry

A simple and highly-secure method for password entry is to type your password whenever an Axelar CLI command prompts for it. For example, you can print the address of your account named `my_account` as follows:

```bash
$AXELARD_HOME/bin/axelard keys show my_account -a
Enter keyring passphrase: {TYPE_YOUR_PASSWORD_HERE}
```

## Automatic password entry

It can be inconvenient to type your password for each Axlear CLI command, especially if you wish to automate CLI commands.

Suppose your keyring password is stored in a shell environment variable called `KEYRING_PASSWORD`. You could prefix your CLI commands with `echo $KEYRING_PASSWORD | `. For example:

```bash
echo $KEYRING_PASSWORD | $AXELARD_HOME/bin/axelard keys show my_account -a
```

<Callout type="error" emoji="☠️">
  Danger: If an attacker were to gain access to your system then the attacker could read your keyring password from your shell environment and then use it to expose your secret keys.
</Callout>

## Axelar documentation elides password entry

For clarity, Axelar CLI documentation elides password entry from CLI commands. You must amend CLI commands according to whichever method of password entry you choose.

Example: to print the address of your account named `my_account` we write only

```bash
$AXELARD_HOME/bin/axelard keys show my_account -a
```
