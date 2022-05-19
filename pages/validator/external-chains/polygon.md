# Polygon

import Callout from 'nextra-theme-docs/callout'

Set up your Polygon Mainnet or Testnet (Mumbai) node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 4-8+ core CPU , 16-32GB RAM, 2TB+ SSD free storage space.
- MacOS or Ubuntu 18.04+
- Build-essential packages
- Golang 1.17+
- [Official Documentation](https://docs.polygon.technology/docs/integrate/full-node-binaries)

## Install required dependencies

In order to build the `polygon` node, you first need to install all of the required dependencies.

### 1. Update and install `build-essential`

```bash
sudo apt-get update
sudo apt-get -y upgrade
sudo apt-get install -y build-essential
```

### 2. Install `golang`

Install the [latest version of golang](https://go.dev/doc/install).

## Install the Polygon node

Polygon node consists of 2 layers, Heimdall and Bor. Heimdall is a fork of tendermint and runs in parallel to the Ethereum network, monitoring contracts, and Bor is a fork of go-Ethereum and producing blocks shuffled by Heimdall nodes. You need to install and run both binaries in the correct order, as explained in the following steps.

### 1. Install Heimdall

Please make sure you checkout the [latest release tag](https://github.com/maticnetwork/heimdall/tags), depending on the network (Mainnet/Testnet). In this tutorial, we are using `v0.2.9`

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

Please make sure you checkout the [latest release tag](https://github.com/maticnetwork/bor/tags). In this tutorial we are using `v0.2.16`

```bash
cd ~/
git clone https://github.com/maticnetwork/bor
cd bor
git checkout v0.2.16
make bor-all
sudo ln -nfs ~/bor/build/bin/bor /usr/bin/bor
sudo ln -nfs ~/bor/build/bin/bootnode /usr/bin/bootnode

# Verify the correct version
bor version
```

## Setup and configure node

### 1. Setup launch directory

Replace the `<network-name>` below with the network you are joining.
Available networks: `mainnet-v1` and `testnet-v4`.

```bash
cd ~/
git clone https://github.com/maticnetwork/launch

mkdir -p node
cp -rf launch/<network-name>/sentry/sentry/* ~/node

# Example for Mainnet:
# cp -rf launch/mainnet-v1/sentry/sentry/* ~/node
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

Again, replace the `<network-name>` below with the network you are joining.

```bash
# Download service file
cd ~/node
wget https://raw.githubusercontent.com/maticnetwork/launch/master/<network-name>/service.sh

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

Open the `~/.heimdalld/config/config.toml` and edit the following flags:

```bash
moniker=<enter unique identifier>

#Mainnet:
seeds="f4f605d60b8ffaaf15240564e58a81103510631c@159.203.9.164:26656,4fb1bc820088764a564d4f66bba1963d47d82329@44.232.55.71:26656,902484e868c6a4bace1bb3cf4b6ba1667561b158@18.228.218.160:26656,afc41bd37d549186cec915c5a4feb3071871cdc1@18.228.98.237:26656,587df41fb0198d72a9e153c608b2c0d840551429@35.182.147.162:26656,ad7bc1c45641454893c74b50357a1bd87778bb50@52.60.36.93:26656"

#Testnet:
seeds="4cd60c1d76e44b05f7dfd8bab3f447b119e87042@54.147.31.250:26656,b18bbe1f3d8576f4b73d9b18976e71c65e839149@34.226.134.117:26656"

Change the value of Pex to true
Change the value of Prometheus to true
Set the max_open_connections value to 100
```

Open the `~/.heimdalld/config/heimdall-config.toml` and edit:

```bash
eth_rpc_url = <insert an RPC endpoint for a fully synced Ethereum mainnet node or Goerli testnet node, i.e Infura.>
```

Open the `~/node/bor/start.sh` and add/change the following flags to start parameters:

```bash
--http --http.addr '0.0.0.0' \

#Mainnet:
--bootnodes "enode://0cb82b395094ee4a2915e9714894627de9ed8498fb881cec6db7c65e8b9a5bd7f2f25cc84e71e89d0947e51c76e85d0847de848c7782b13c0255247a6758178c@44.232.55.71:30303,enode://88116f4295f5a31538ae409e4d44ad40d22e44ee9342869e7d68bdec55b0f83c1530355ce8b41fbec0928a7d75a5745d528450d30aec92066ab6ba1ee351d710@159.203.9.164:30303"

#Testnet:
--bootnodes "enode://320553cda00dfc003f499a3ce9598029f364fbb3ed1222fdc20a94d97dcc4d8ba0cd0bfa996579dcc6d17a534741fb0a5da303a90579431259150de66b597251@54.147.31.250:30303"
```

To enable Archive mode you can also add the following flags in the start.sh file
```bash
--gcmode 'archive' \
--ws --ws.port 8546 --ws.addr 0.0.0.0 --ws.origins '*' \
```


### 5. Download maintained snapshots

<Callout emoji="ℹ️">
  Syncing Heimdall and Bor services can take several days to sync fully. Alternatively, you can use snapshots to reduce the sync time to a few hours. If you wish to sync the node from the start, then you can skip this step.
</Callout>

To use the snapshots, please visit [Polygon Chains Snapshots](https://snapshots.matic.today/) and download the latest available snapshot for Heimdall and Bor. Replace the `snapshot-link` below with the full path to the snapshot of the network you're joining.

```bash
wget <snapshot-link-heimdall> -O - | tar -xzf - -C ~/.heimdalld/data/
wget <snapshot-link-bor> -O - | tar -xzf - -C ~/.bor/data/bor/chaindata
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

Once `heimdalld` is synced with the [latest block height](https://wallet.polygon.technology/staking/), then you can start the `bor` service file:

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
