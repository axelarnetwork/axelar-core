# Ethereum

import Callout from 'nextra-theme-docs/callout'
import CodeBlock from '../../../components/code-block'
import Tabs from '../../../components/tabs'

Set up your Ethereum Mainnet or Goerli Testnet RPC node

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: CPU with 2+ cores, 4GB RAM, 600GB+ free storage space.
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://ethereum.org/en/developers/docs/nodes-and-clients)


## Steps
1. Install execution client
2. Install consensus layer client
3. Configure systemd
4. Configure vald


### Install Execution client

Install an [execution client](https://ethereum.org/en/developers/docs/nodes-and-clients/#execution-clients) for Ethereum.
In this sample guide we will be installing `Geth` with the built-in launchpad PPAs (Personal Package Archives) on Ubuntu. If you are on different OS, please refer to the [official Documentation](https://geth.ethereum.org/docs/getting-started).

```bash
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install ethereum
```

### Install a consensus layer client

To sync after the latest merge in Goerli network geth nodes should run a consensus client to be able to keep in sync with the chain. The list of consensus clients can be found in https://ethereum.org/en/developers/docs/nodes-and-clients/#consensus-clients

Users can opt for any Consensus client. The sample instructions below
use Prysm as the Consensus client.
Please refer to the [Prysm docs](https://docs.prylabs.network/docs/install/install-with-script) for the up-to-date instructions. For fast syncing,
consider [checkpoint sync](https://docs.prylabs.network/docs/prysm-usage/checkpoint-sync) instead of genesis sync.

```bash
mkdir prysm && cd prysm

# Download installation script
curl https://raw.githubusercontent.com/prysmaticlabs/prysm/master/prysm.sh --output prysm.sh && chmod +x prysm.sh

# Download the Goerli network genesis file
wget https://github.com/eth-clients/eth2-networks/raw/master/shared/prater/genesis.ssz

# Generate a 32 byte hex secret key for authentication, for e.g.
./prysm.sh beacon-chain generate-auth-secret
# Alternatively, `openssl rand -hex 32 | tr -d "\n" > "jwt.hex"`

# Start Prysm beacon chain.
# NOTE: For Goerli testnet, add `--prater` flag
./prysm.sh beacon-chain --http-web3provider=http://localhost:8551 --jwt-secret=/path/to/jwt.hex --genesis-state=./genesis.ssz
```

### Configure systemd

##### 1. Create systemd service file

After installation of `go-ethereum`, we are now ready to start the `geth` process but in order to ensure it is running in the background and auto-restarts in case of a server failure, we will setup a service file with systemd.

<Callout type="error" emoji="⚠️">
  Replace `$USER` and path to `geth` in the config below.
  For Goerli Testnet, add the `--goerli` flag to the `geth` command.
</Callout>

Add the following to `/etc/systemd/system/geth.service`
```yaml
[Unit]
Description=Ethereum Node
After=network.target

[Service]
User=$USER
Type=simple
ExecStart=/usr/bin/geth --syncmode "snap" --http --http.api=eth,net,web3,engine --http.vhosts * --http.addr 0.0.0.0 --authrpc.jwtsecret=/path/to/jwt.hex --override.terminaltotaldifficulty 50000000000000000
Restart=on-failure
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

##### 2. Start `geth`

```bash
sudo systemctl enable geth
sudo systemctl daemon-reload
sudo systemctl start geth
```

If everything was set-up correctly, your Ethereum node should now be starting the process of synchronization. This will take several hours, depending on your hardware.To check the status of the running service or to follow the logs, you can use:

```bash
sudo systemctl status geth
sudo journalctl -u geth -f
```

##### 3. Verify sync status

To verify if your node is in sync you can check the latest block from the [explorer](https://goerli.etherscan.io/).
Compare it with what you have in `sudo journalctl -u geth -f`


##### 4. Test RPC connection

Alternatively, you can now also use the Geth JavaScript console and check status of your node by attaching to your newly created `geth.ipc`. Don't forget to replace `$USER` and `path`, depending on your server configuration.

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`geth attach ipc:/root/.ethereum/geth.ipc
eth.syncing`}
</CodeBlock>
},
{
title: "Goerli Testnet",
content: <CodeBlock language="bash">
{`geth attach ipc:/root/.ethereum/goerli/geth.ipc
eth.syncing`}
</CodeBlock>
}
]} />

Once your node is fully synced, the output from above will say `false`. To test your Ethereum RPC node, you can send an RPC request using `cURL`

```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
```

If you are testing it remotely, please replace `localhost` with the IP or URL of your server.

### Configure vald

In order for `vald` to connect to your Ethereum node, your `rpc_addr` should be exposed in
vald's `config.toml`

<Callout emoji="📝">
  Goerli testnet chain name is `ethereum-2`
</Callout>

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="yaml">
{`[[axelar_bridge_evm]]
name = "Ethereum"
rpc_addr = "http://IP:PORT"
start-with-bridge = true`}
</CodeBlock>
},
{
title: "Goerli Testnet",
content: <CodeBlock language="yaml">
{`[[axelar_bridge_evm]]
name = "ethereum-2"
rpc_addr = "http://IP:PORT"
start-with-bridge = true`}
</CodeBlock>
}
]} />


### Upgrade geth

If new release of geth is out, you can update it via the following.
Check that you are on [latest version](https://github.com/ethereum/go-ethereum/releases).

```bash
geth version

# Upgrade geth if version is outdated
sudo systemctl stop geth
sudo apt-get update
sudo apt-get upgrade geth

# Start you `geth` service and check logs.
sudo systemctl enable geth
sudo systemctl start geth
journalctl -u geth -f -n 100
```
