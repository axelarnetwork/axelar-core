# Avalanche

import Callout from 'nextra-theme-docs/callout'

Set up your Avalanche Fuji Testnet node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 8 AWS vCPU+, 16GB RAM, 100GB+ free storage space.
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://docs.avax.network/build/tutorials/nodes-and-staking/run-avalanche-node)

## Install AvalancheGo

In this guide we will be using a bash script created by the Avalanche team, which will automatically set up a running node with minimum user input required.

##### 1. Download `avalanchego-installer`

```bash
wget -nd -m https://raw.githubusercontent.com/ava-labs/avalanche-docs/master/scripts/avalanchego-installer.sh
```

##### 2. Set permission and run the script

```bash
chmod 755 avalanchego-installer.sh
./avalanchego-installer.sh
```

The script will start and prompt you for information about your server environment. Follow the required steps, enter your network information and confirm the installation. The script will then create and start the `avalanchego.service` for you automatically. To check if the service is running or follow the logs, use the following commands:

```bash
sudo systemctl status avalanchego
sudo journalctl -u avalanchego -f
```

<Callout type="error" emoji="☠️">
  Danger: By default the network will start synchronizing on the Mainnet but we want to run our node on the Avalanche Fuji testnet, so you need to stop the `avalanchego.service`, edit the `node.json` configuration file and restart the service.
</Callout>

```bash
sudo systemctl stop avalanchego
nano  /home/avax/.avalanchego/configs/node.json
```

Change network-id to `"fuji"`, save the file and restart the service:

```bash
sudo systemctl start avalanchego
```

Now you should be synchronizing on Fuji testnet network. Once the network is fully synced, you should see a message like:

`health/service.go#130: "isBootstrapped" became healthy with: {"timestamp":"2021-12-27T21:35:36.879654389+01:00","duration":6943,"contiguousFailures":0,"timeOfFirstFailure":null}`

## Test your Avalanche RPC connection

Once your node is fully synced, you can run a cURL request to see the status of your node:

```bash
curl -X POST --data '{"jsonrpc": "2.0","method": "info.isBootstrapped","params":{"chain":"C"},"id":1}' -H 'content-type:application/json;' localhost:9650/ext/info

```

If the node is successfully synced, the output from above will print `{"jsonrpc":"2.0","result":{"isBootstrapped":true},"id":1}`

#### EVM RPC endpoint URL

Axelar Network will be connecting to the C-Chain, which is the EVM compatible blockchain, so your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT/ext/bc/C/rpc
```

Example:
`http://192.168.192.168:9650/ext/bc/C/rpc`
