# Overview

import Callout from 'nextra-theme-docs/callout'

An Axelar network validator participates in block creation, multi-party cryptography protocols, and voting.

Convert your existing Axelar node into a validator by staking AXL tokens and attaching external EVM-compatible blockchains.

## Prerequisites

- **Hardware:** Minimum: 16 cores, 16GB RAM, 1.5 TB drive. Recommended: 32 cores, 32 GB RAM, 2 TB+ drive.
- You have downloaded the Axelar blockchain and are comfortable with [Basic node management](../../node/basic).

MOVE:

- Your Axelar node has an account named `validator` that you control. Let `{VALIDATOR_ADDR}` denote the address of your `validator` account.
- Backup your `validator` secret mnemonic and your Tendermint consensus secret key as per [Quick sync](../node/join).

## Steps to become a validator

1. [Configure companion processes](config)
2. [Create and backup accounts](backup)
3. [Launch companion processes](vald-tofnd)
4. [Register broadcaster proxy](register-broadcaster)
5. [Stake AXL tokens on the Axelar network](stake-axl)
6. [Health check](health-check)
7. [Set up external chains](../external-chains)

## Other setup-related tasks

- [Troubleshoot start-up](../troubleshoot/startup)
- [Recover validator from mnemonic or secret keys](../troubleshoot/recovery)
- [Leave as a validator](../troubleshoot/leave)
- [Missed too many blocks](../troubleshoot/missed-too-many-blocks)
