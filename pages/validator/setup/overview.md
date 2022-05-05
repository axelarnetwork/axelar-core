# Overview

import Callout from 'nextra-theme-docs/callout'

An Axelar network validator participates in block creation, multi-party cryptography protocols, and voting.

Convert your existing Axelar node into a validator by staking AXL tokens and attaching external EVM-compatible blockchains.

<Callout type="error" emoji="🔥">
  Important: The Axelar network is under active development. Use at your own risk with funds you're comfortable using. See [Terms of use](/terms-of-use).
</Callout>

## Prerequisites

- **Hardware:** Minimum: 16 cores, 16GB RAM, 1.5 TB drive. Recommended: 32 cores, 32 GB RAM, 2 TB+ drive.
- You have downloaded the Axelar blockchain and are comfortable with [Basic node management](../node/basic).
- Your Axelar node has an account named `validator` that you control. Let `{VALIDATOR_ADDR}` denote the address of your `validator` account.
- Backup your `validator` secret mnemonic and your Tendermint consensus secret key as per [Quick sync](../node/join).
- You have configured your environment for `axelard` CLI commands as per [Configure your environment](../../node/config).

## Steps to become a validator

1. [Launch companion processes for the first time](../setup/vald-tofnd)
2. [Back-up your validator mnemonics and secret keys](../setup/backup)
3. [Register broadcaster proxy](../setup/register-broadcaster)
4. [Stake AXL tokens on the Axelar network](../setup/stake-axl)
5. [Health check](./health-check)
6. [Set up external chains](../external-chains)

## Other setup-related tasks

- [Troubleshoot start-up](../troubleshoot/startup)
- [Recover validator from mnemonic or secret keys](../troubleshoot/recovery)
- [Leave as a validator](../troubleshoot/leave)
- [Missed too many blocks](../troubleshoot/missed-too-many-blocks)
