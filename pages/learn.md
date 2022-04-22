# Learn about Axelar

import Button from '../components/button'

UNDER CONSTRUCTION

## Gateway smart contracts

What is a gateway contract and what does it do? Each external chain has its own Axelar gateway contract. Validators hold key shares that allow them to do things on these chains via gateway txs. Ex: mint/burn tokens, approve GMP txs.

## Validators

What do Axelar validators do? Hold key shares to gateway contracts, vote on external-chain events, use gateway key shares to authorize actions on external chains

## Microservices

What is MS and what does it do? MS is an optional convenience provided by Axelar. MS does tasks that can be performed by anyone (ie. no need for trust) but must be done by at least someone. Eg. listen for an event on external chains and propose a validator vote.

## Gas receiver

Gas receiver is an example of MS: without gas receiver your GMP tx is “approved” but not “executed”. You could execute it yourself or you could use Axelar’s gas receiver to automatically execute your approved tx.

## AxelarJS SDK

[link](./learn/sdk)

## CLI

[link](./learn/cli)
