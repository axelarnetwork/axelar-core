# Ethereum

import Callout from 'nextra-theme-docs/callout'

Set up your Ethereum Mainnet or Ropsten Testnet node post transition to PoS

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: CPU with 2+ cores, 4GB RAM, 600GB+ free storage space.
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://geth.ethereum.org/docs/getting-started)

# High level Steps
1. Install Geth
2. Install consensus layer
3. Expose methods


## Install Geth

In this guide we will be installing `Geth` with the built-in launchpad PPAs (Personal Package Archives) on Ubuntu. If you are on different OS, please refer to the [official Documentation](https://geth.ethereum.org/docs/getting-started).

Note: For post merge sync Geth version should be minimum 1.10.18-stable

##### 1. Enable launchpad repository

```bash
sudo add-apt-repository -y ppa:ethereum/ethereum
```

##### 2. Install the latest version of go-ethereum:

```bash
sudo apt-get update
sudo apt-get install ethereum
```

##### 3. Install a consensus layer :

To sync after the latest merge in Ropsten network geth nodes should run a consensus client to be able to keep in sync with the chain. The list of consensus clients can be found in https://ethereum.org/en/developers/docs/nodes-and-clients/#consensus-clients

Users can opt for any Consensus client. To install Prysm as Consensus client 

```bash
mkdir prysm && cd prysm
curl https://raw.githubusercontent.com/prysmaticlabs/prysm/master/prysm.sh --output prysm.sh && chmod +x prysm.sh
export USE_PRYSM_VERSION=v2.1.3-rc.3
#Download the Ropsten network genesis file
wget https://github.com/eth-clients/merge-testnets/raw/main/ropsten-beacon-chain/genesis.ssz
#Generate a secret key for authentication 
Generate a random 32byte hex string and store it in a local directory 
#Start Local Prysm beacon chain 
./prysm.sh beacon-chain --http-web3provider=http://localhost:8551  --jwt-secret=/PathToFile/jwtsecret --ropsten --genesis-state=./genesis.ssz --block-batch-limit=64
```

Refer:

https://docs.prylabs.network/docs/execution-node/authentication/

https://docs.prylabs.network/docs/next/install/install-with-script

[https://seanwasere.com/generate-random-hex/]

## Run `geth` through systemd

##### 1. Create systemd service file

After installation of `go-ethereum`, we are now ready to start the `geth` process but in order to ensure it is running in the background and auto-restarts in case of a server failure, we will setup a service file with systemd.

<Callout emoji="📝">
  Note: In the service file below you need to replace `$USER` and path to `geth`, depending on your system configuration.
</Callout>

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/geth.service
[Unit]
Description=Ethereum Node
After=network.target

[Service]
User=$USER
Type=simple
ExecStart=/usr/bin/geth --syncmode "snap" --http --http.api=eth,net,web3,engine --http.vhosts * --http.addr 0.0.0.0 --authrpc.jwtsecret=/PathToFile/jwtsecret --override.terminaltotaldifficulty 50000000000000000
Restart=on-failure
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```
<Callout type="error" emoji="⚠️">
 If you would like to run a node on the Testnet instead (Ropsten), you need to add the `--ropsten` flag to the configuration above. 
</Callout>

##### 2. Enable and start the `geth` service

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

##### 3. Verify if its in sync

To verify if your node is in sync you can check the latest block from the [explorer](https://ropsten.etherscan.io/). 
Compare it with what you have in `sudo journalctl -u geth -f` 


## Test your Ethereum RPC connection

Alternatively, you can now also use the Geth JavaScript console and check status of your node by attaching to your newly created `geth.ipc`. Don't forget to replace $USER and path, depending on your server configuration.

```bash
geth attach ipc:/root/.ethereum/geth.ipc
eth.syncing

#Testnet
#geth attach ipc:/root/.ethereum/ropsten/geth.ipc
```

Once your node is fully synced, the output from above will say `false`. To test your Ethereum node, you can send an RPC request using `cURL`

```bash
curl -X POST http://localhost:8545 \
-H "Content-Type: application/json" \
--data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
```

If you are testing it remotely, please replace `localhost` with the IP or URL of your server.

#### EVM RPC endpoint URL

In order for Axelar Network to connect to your Ethereum node, your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:8545`

### Upgrade geth if necessary

If new release of geth is out, you can update it by following commands.

```
sudo systemctl stop geth
sudo apt-get update
sudo apt-get upgrade geth
```
Check that you are on latest version. The last version is `1.10.20-stable`
```
geth version
```
Start you `geth` service and check logs.
```
sudo systemctl enable geth
sudo systemctl start geth
journalctl -u geth -f -n 100
```
