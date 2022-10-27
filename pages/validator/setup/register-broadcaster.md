# Register broadcaster proxy

import Callout from 'nextra-theme-docs/callout'

Axelar validators exchange messages with one another via the Axelar blockchain. Each validator sends these messages from a separate `broadcaster` account.

<Callout type="warning" emoji="⚠️">
  Caution: A validator can only register one `broadcaster` address throughout its lifetime. This `broadcaster` address cannot be changed after it has been registered. If you need to register a different proxy address then you must also create an entirely new validator.
</Callout>

## Learn your broadcaster account address

```bash
$AXELARD_HOME/bin/axelard keys show broadcaster -a --home $AXELARD_HOME
```

Let `{BROADCASTER_ADDR}` denote your `broadcaster` address

## Fund your validator and broadcaster accounts

**Testnets:**
Go to the Axelar testnet faucet and send some free AXL testnet tokens to both `{BROADCASTER_ADDR}` and `{VALIDATOR_ADDR}`:

- [Testnet-1 Faucet](https://faucet.testnet.axelar.dev/).
- [Testnet-2 Faucet](https://faucet-casablanca.testnet.axelar.dev/)

## Register your broadcaster account

```bash
$AXELARD_HOME/bin/axelard tx snapshot register-proxy {BROADCASTER_ADDR} --from validator --chain-id $AXELARD_CHAIN_ID --home $AXELARD_HOME --gas auto --gas-adjustment 1.4
```

## Optional: check your broadcaster registration

```bash
$AXELARD_HOME/bin/axelard q snapshot proxy {VALOPER_ADDR}
```
