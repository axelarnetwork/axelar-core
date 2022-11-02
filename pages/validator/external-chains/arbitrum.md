# Arbitrum

import Callout from 'nextra-theme-docs/callout'
import CodeBlock from '../../../components/code-block'
import Tabs from '../../../components/tabs'

Set up your Arbitrum Mainnet & Goerli Testnet RPC node

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- [Setup your Ethereum node](ethereum/)
- Minimum hardware requirements: CPU with 2+ cores, 4GB RAM, 600GB+ free storage space.
- MacOS or Ubuntu 20.04+
- [Official Documentation](https://developer.offchainlabs.com/node-running/running-a-node)


## Steps
1. Install Docker
2. Install Arbitrum image
3. Configure vald


### Install docker

```bash
sudo apt update && sudo apt install curl jq -y < "/dev/null"
curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh
```

### Install Arbitrum node

```bash
mkdir -p $HOME/data/arbitrum
chmod -fR 777 $HOME/data/arbitrum
```
Now, you will see this flag in the command below `--l1.url <YOUR_ETH_RPC_URL>` this means that your arbitrum node needs a synced Ethereum node.
Please provide the RPC URL of a synced Ethereum node with this flag.

<Callout type="error" emoji="⚠️">
  Please avoid using 3rd party providers like alchemy, infura etc. These providers have a specific request limit, and your node can throw 100s of thousands of requests while trying to sync.
</Callout>

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`docker run --rm -it -d -v /path/to/data/arbitrum:/home/user/.arbitrum -p 0.0.0.0:8547:8547 -p 0.0.0.0:8548:8548 offchainlabs/nitro-node:v2.0.7-10b845c --l1.url <YOUR_ETH_RPC_URL> --l2.chain-id=42161 --http.api=net,web3,eth,debug --http.corsdomain=* --http.addr=0.0.0.0 --http.vhosts=* --init.url="https://snapshot.arbitrum.io/mainnet/nitro.tar"`}
</CodeBlock>
},
{
title: "Goerli Testnet",
content: <CodeBlock language="bash">
{`docker run --rm -it -d -v /path/to/data/arbitrum:/home/user/.arbitrum -p 0.0.0.0:8547:8547 -p 0.0.0.0:8548:8548 offchainlabs/nitro-node:v2.0.7-10b845c --l1.url <YOUR_GOERLI_ETH_RPC_URL> --l2.chain-id=421613 --http.api=net,web3,eth,debug --http.corsdomain=* --http.addr=0.0.0.0 --http.vhosts=*`}
</CodeBlock>
}
]} />

#### Verify sync status

To verify if your node is in sync you can check the latest block from the explorer.
Compare it with what you have in `docker ps -q | xargs -L 1 docker logs --tail 10 -f`

- [Mainnet Explorer](https://arbiscan.io)
- [Testnet Explorer](https://goerli.arbiscan.io)

#### Test RPC connection

Once your node is fully synced, the output from above will say `false`. To test your Arbitrum RPC node, you can send an RPC request using `cURL`

```bash
curl -X POST http://localhost:8547 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
```

If you are testing it remotely, please replace `localhost` with the IP or URL of your server.

### Configure vald

In order for `vald` to connect to your Arbitrum node, your `rpc_addr` should be exposed in
vald's `config.toml`


<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="yaml">
{`[[axelar_bridge_evm]]
name = "arbitrum"
l1_chain_name = "Ethereum"
rpc_addr = "http://IP:PORT"
start-with-bridge = true`}
</CodeBlock>
},
{
title: "Goerli Testnet",
content: <CodeBlock language="yaml">
{`[[axelar_bridge_evm]]
name = "arbitrum"
l1_chain_name = "ethereum-2"
rpc_addr = "http://IP:PORT"
start-with-bridge = true`}
</CodeBlock>
}
]} />
