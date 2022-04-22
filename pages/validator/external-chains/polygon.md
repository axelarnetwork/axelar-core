# Polygon

import Callout from 'nextra-theme-docs/callout'

Set up your Polygon Mumbai Testnet node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 4-8+ core CPU , 16-32GB RAM, 100GB+ free storage space.
- MacOS or Ubuntu 18.04+
- Build-essential packages
- Golang 1.17+
- [Official Documentation](https://docs.polygon.technology/docs/integrate/full-node-binaries)

## Install required dependencies

In order to build the `go-opera`, you first need to install all of the required dependencies.

### 1. Update and install `build-essential`

```bash
sudo apt-get update
sudo apt-get -y upgrade
sudo apt-get install -y build-essential
```

### 2. Install `golang`

Install the [latest version of golang](https://go.dev/doc/install).

### 2. Install `RabbitMq`

```bash
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.9-management
```

## Install the Polygon node

Polygon node consists of 2 layers, `Heimdall` and `Bor`. Heimdall is a fork of tendermint and is running in parallel to an Ethereum network, monitoring contracts. Bor is a fork of go-Ethereum and producing blocks shuffled by Heimdall nodes. You need to install and run both binaries in the correct order as explained in the following steps.

### 1. Install Heimdall

Please make sure you checkout the [latest release tag](https://github.com/maticnetwork/heimdall/tags). In this tutorial we are using `v0.2.9`

```bash
cd ~/
git clone https://github.com/maticnetwork/heimdall
cd heimdall
git checkout v0.2.9
make install

# Verify the correct version
heimdalld version --long
```

### 2. Install Bor

Please make sure you checkout the [latest release tag](https://github.com/maticnetwork/bor/tags). In this tutorial we are using `v0.2.14`

```bash
cd ~/
git clone https://github.com/maticnetwork/bor
cd bor
git checkout v0.2.14
make bor-all
sudo ln -nfs ~/bor/build/bin/bor /usr/bin/bor
sudo ln -nfs ~/bor/build/bin/bootnode /usr/bin/bootnode

# Verify the correct version
bor version
```

## Setup and configure node

### 1. Setup launch directory

```bash
cd ~/
git clone https://github.com/maticnetwork/launch

mkdir -p node
cp -rf launch/testnet-v4/sentry/sentry/* ~/node
```

### 2. Setup network directories

```bash
# Heimdall
cd ~/node/heimdall
bash setup.sh

# Bor
cd ~/node/bor
bash setup.sh
```

### 3. Setup service files

```bash
# Download service file
cd ~/node
wget https://raw.githubusercontent.com/maticnetwork/launch/master/testnet-v4/service.sh

# Generate Metadata
sudo mkdir -p /etc/matic
sudo chmod -R 777 /etc/matic/
touch /etc/matic/metadata

# Generate service file and copy them into systemd directory
cd ~/node
bash service.sh
sudo cp *.service /etc/systemd/system/
```

### 4. Setup config files

Open the `~/.heimdalld/config/config.toml` and edit:

```bash
moniker=<enter unique identifier>
seeds="4cd60c1d76e44b05f7dfd8bab3f447b119e87042@54.147.31.250:26656,b18bbe1f3d8576f4b73d9b18976e71c65e839149@34.226.134.117:26656"
```

Open the `~/.heimdalld/config/heimdall-config.toml` and edit:

```bash
eth_rpc_url = <insert Infura or any full node RPC URL to Goerli>
```

Open the `~/node/bor/start.sh` and add the following flag to start parameters:

```bash
--bootnodes "enode://320553cda00dfc003f499a3ce9598029f364fbb3ed1222fdc20a94d97dcc4d8ba0cd0bfa996579dcc6d17a534741fb0a5da303a90579431259150de66b597251@54.147.31.250:30303"
```

### 5. Download maintained snapshots

<Callout emoji="ℹ️">
  Info: Syncing Heimdall and Bor services can take several days to fully sync. Alternatively, you can use snapshots which will reduce the sync time to few hours. If you wish to sync the node from start, then you can skip this step.
</Callout>

In order to use the snapshots, please visit [Polygon Chains Snapshots](https://snapshots.matic.today/) and download the latest available snapshots fot Heimdall and Bor. In this guide we are using:

```bash
wget https://matic-blockchain-snapshots.s3-accelerate.amazonaws.com/matic-mumbai/heimdall-snapshot-2021-12-09.tar.gz -O - | tar -xzf - -C ~/.heimdalld/data/
wget https://matic-blockchain-snapshots.s3-accelerate.amazonaws.com/matic-mumbai/bor-fullnode-node-snapshot-2021-12-15.tar.gz -O - | tar -xzf - -C ~/.bor/data/bor/chaindata
# If needed, change the path depending on your server configuration.
```

## Start the Polygon services

After completing all of the previous steps, your node should be configured and ready to launch with the previously created service files.

### 1. Start Heimdalld

```bash
sudo service heimdalld start
sudo service heimdalld-rest-server start
```

<Callout type="warning" emoji="⚠️">
  Important: You need to wait for Heimdall node to fully sync with the network before starting the Bor service!
</Callout>

You can check the status of `heimdalld` service or follow the logs with:

```bash
sudo service heimdalld status
journalctl -u heimdalld.service -f
```

### 2. Start Bor

Once `heimdalld` is synced with the [latest block height](https://wallet-dev.polygon.technology/staking/), then you can start the `bor` service file:

```bash
sudo service bor start

# Check status and logs
sudo service bor status
journalctl -u bor.service -f
```

## Test your Polygon RPC connection

Once your `Bor` node is fully synced, you can run a cURL request to see the status of your node:

```bash
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "eth_syncing", "params":[]}' localhost:8545
```

If the node is successfully synced, the output from above will print `{"jsonrpc":"2.0","id":1,"result":false}`

### EVM RPC endpoint URL

Axelar Network will be connecting to the EVM compatible blockchain `Bor`, so your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:8545`
