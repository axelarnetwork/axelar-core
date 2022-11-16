# CELO

import Callout from 'nextra-theme-docs/callout'
import CodeBlock from '../../../components/code-block'
import Tabs from '../../../components/tabs'

Set up your CELO Mainnet or Alfajores Testnet RPC node

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- [Minimum hardware requirements](https://docs.celo.org/validator/run/mainnet#hardware-requirements):
   - Memory: 8 GB RAM
   - CPU: Quad core 3GHz (64-bit)
   - Disk: 256 GB of SSD storage, plus a secondary HDD desirable
   - Network: At least 1 GB input/output Ethernet with a fiber Internet connection, ideally redundant connections and HA switches
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://docs.celo.org/network/node/run-mainnet)


## Steps
1. Setup Preferences
2. Celo Networks
3. Pull the Celo Docker image
4. Set up a data directory
5. Create an account and get its address
6. Start the node
7. Check Synced
8. Configure vald
9. Upgrade Celo


### Setup Preferences

Update and upgrade the packages by running the following command in the terminal:
```bash
sudo apt-get update && sudo apt-get upgrade
```

Install Required Packages:
```bash
sudo apt-get install docker.io
```

### Celo Networks

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
`export CELO_IMAGE=us.gcr.io/celo-org/geth:mainnet`
</CodeBlock>
},
{
title: "Alfajores Testnet",
content: <CodeBlock language="bash">
`export CELO_IMAGE=us.gcr.io/celo-org/geth:alfajores`
</CodeBlock>
}
]} />

### Pull the Celo Docker image

We're going to use a Docker image containing the Celo node software in this tutorial.

If you are re-running these instructions, the Celo Docker image may have been updated, and it's important to get the latest version.
```bash
docker pull $CELO_IMAGE
```

### Set up a data directory

First, create the directory that will store your node's configuration and its copy of the blockchain. This directory can be named anything you'd like, but here's a default you can use. The commands below create a directory and then navigate into it. The rest of the steps assume you are running the commands from inside this directory.
```bash
mkdir -r ~/celo-data-dir
cd ~/celo-data-dir
export CELO_DIR=~/celo-data-dir
```

### Create an account and get its address
In this step, you'll create an account on the network. If you've already done this and have an account address, you can skip this and move on to configuring your node.

Run the command to create a new account:
```bash
docker run -v $CELO_DIR:/root/.celo --rm -it $CELO_IMAGE account new
```

Example Result:
```bash
INFO [10-28|11:23:37.486] Maximum peer count                       ETH=175 LES=0 total=175
Your new account is locked with a password. Please give a password. Do not forget this password.
Password:
Repeat password:

Your new key was generated

Public address of the key:   <YOUR-ACCOUNT-ADDRESS>
Path of the secret key file: /root/.celo/keystore/UTC--2022-10-28T11-23-45.863789512Z--2aa3b36ff21ecda6d7b277c730e6ef4f7e173598

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
```

Save this address to an environment variables, so that you can reference it below (don't include the braces):
```bash
export CELO_ACCOUNT_ADDRESS=<YOUR-ACCOUNT-ADDRESS>
```

This environment variable will only persist while you have this terminal window open.
Add it to `~/.bash_profile` for future use.
```bash
echo "export CELO_IMAGE=$CELO_IMAGE" >> ~/.bash_profile
echo "export CELO_ACCOUNT_ADDRESS=$CELO_ACCOUNT_ADDRESS" >> ~/.bash_profile
echo "export CELO_DIR=$CELO_DIR" >> ~/.bash_profile
```

### Start the node

This command specifies the settings needed to run the node, and gets it started.

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
`docker run --name celo-fullnode -d --restart unless-stopped --stop-timeout 300 -p 8545:8545 -p 8546:8546 -p 30303:30303 -p 30303:30303/udp -v $CELO_DIR:/root/.celo $CELO_IMAGE --verbosity 3 --syncmode full --http --http.addr 0.0.0.0 --http.api eth,net,web3,debug,admin,personal --light.serve 90 --light.maxpeers 1000 --maxpeers 1100 --etherbase $CELO_ACCOUNT_ADDRESS --datadir /root/.celo`
</CodeBlock>
},
{
title: "Alfajores Testnet",
content: <CodeBlock language="bash">
`docker run --name celo-fullnode -d --restart unless-stopped --stop-timeout 300 -p 8545:8545 -p 8546:8546 -p 30303:30303 -p 30303:30303/udp -v $CELO_DIR:/root/.celo $CELO_IMAGE --verbosity 3 --syncmode full --http --http.addr 0.0.0.0 --http.api eth,net,web3,debug,admin,personal --light.serve 90 --light.maxpeers 1000 --maxpeers 1100 --etherbase $CELO_ACCOUNT_ADDRESS --alfajores --datadir /root/.celo`
</CodeBlock>
}
]} />

You'll start seeing some output. After a few minutes, you should see lines that look like this. This means your node has started syncing with the network and is receiving blocks.
```bash
INFO [11-03|07:09:49.666] Imported new chain segment               blocks=1  txs=25  mgas=3.463  elapsed=48.153ms    mgasps=71.914  number=15,958,278 hash=edfe6c..2bb604 dirty=156.40MiB
INFO [11-03|07:09:54.584] Imported new chain segment               blocks=1  txs=23  mgas=2.541  elapsed=40.803ms    mgasps=62.268  number=15,958,279 hash=0b359e..9d77ab dirty=156.34MiB
INFO [11-03|07:09:59.690] Imported new chain segment               blocks=1  txs=37  mgas=4.989  elapsed=61.154ms    mgasps=81.576  number=15,958,280 hash=3b8915..b2506f dirty=156.39MiB
INFO [11-03|07:10:04.615] Imported new chain segment               blocks=1  txs=26  mgas=2.598  elapsed=41.537ms    mgasps=62.537  number=15,958,281 hash=36f20e..4afcee dirty=156.33MiB
INFO [11-03|07:10:09.864] Imported new chain segment               blocks=1  txs=36  mgas=5.290  elapsed=79.355ms    mgasps=66.663  number=15,958,282 hash=61906a..1519fe dirty=156.36MiB
INFO [11-03|07:10:14.669] Imported new chain segment               blocks=1  txs=38  mgas=4.118  elapsed=54.928ms    mgasps=74.978  number=15,958,283 hash=af41a9..f285e2 dirty=156.41MiB
INFO [11-03|07:10:19.736] Imported new chain segment               blocks=1  txs=20  mgas=4.494  elapsed=51.821ms    mgasps=86.713  number=15,958,284 hash=e8bd7c..30260b dirty=156.44MiB
INFO [11-03|07:10:24.684] Imported new chain segment               blocks=1  txs=28  mgas=4.218  elapsed=51.408ms    mgasps=82.054  number=15,958,285 hash=c1a6be..c7825e dirty=156.42MiB
```

<Callout type="error" emoji="⚠️">
Security: The command line above includes the parameter --http.addr 0.0.0.0 which makes the Celo Blockchain software listen for incoming RPC requests on all network adaptors. Exercise extreme caution in doing this when running outside Docker, as it means that any unlocked accounts and their funds may be accessed from other machines on the Internet. In the context of running a Docker container on your local machine, this together with the docker -p flags allows you to make RPC calls from outside the container, i.e from your local host, but not from outside your machine. Read more about Docker Networking here.
</Callout>

### Check Synced

Once your node is fully synced, the output from above will say `false`. To test your Celo RPC node, you can send an RPC request using `cURL`
```bash
curl https://localhost:8545 \
--request POST \
--header "Content-Type: application/json" \
--data '{ "jsonrpc":"2.0", "method":"eth_blockNumber","params":[],"id":1}'
```

### Configure vald

In order for `vald` to connect to your Ethereum node, your `rpc_addr` should be exposed in
vald's `config.toml`

```bash
[[axelar_bridge_evm]]
name = "celo"
rpc_addr = "<node-rpc-addr>"
start-with-bridge = true
```

### Upgrade Celo

#### Recent Releases

- [You can view the latest releases here.](https://github.com/celo-org/celo-blockchain/releases)

#### Pull the latest Docker image
```bash
docker pull $CELO_IMAGE
```

#### Stop and remove the existing node
```bash
docker stop -t 300 celo-fullnode
docker rm celo-fullnode
```

#### Start the new node

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
`docker run --name celo-fullnode -d --restart unless-stopped --stop-timeout 300 -p 8545:8545 -p 8546:8546 -p 30303:30303 -p 30303:30303/udp -v $CELO_DIR:/root/.celo $CELO_IMAGE --verbosity 3 --syncmode full --http --http.addr 0.0.0.0 --http.api eth,net,web3,debug,admin,personal --light.serve 90 --light.maxpeers 1000 --maxpeers 1100 --etherbase $CELO_ACCOUNT_ADDRESS --datadir /root/.celo`
</CodeBlock>
},
{
title: "Alfajores Testnet",
content: <CodeBlock language="bash">
`docker run --name celo-fullnode -d --restart unless-stopped --stop-timeout 300 -p 8545:8545 -p 8546:8546 -p 30303:30303 -p 30303:30303/udp -v $CELO_DIR:/root/.celo $CELO_IMAGE --verbosity 3 --syncmode full --http --http.addr 0.0.0.0 --http.api eth,net,web3,debug,admin,personal --light.serve 90 --light.maxpeers 1000 --maxpeers 1100 --etherbase $CELO_ACCOUNT_ADDRESS --alfajores --datadir /root/.celo`
</CodeBlock>
}
]} />
