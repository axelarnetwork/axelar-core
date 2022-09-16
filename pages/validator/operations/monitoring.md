import Callout from 'nextra-theme-docs/callout'

# Monitoring a Validator or Node

There are a few different aspects to monitoring your validator. We need to monitor both the infrastructure and the axelard & supporting processes.

## Infrastructure
Infrastructure setups differ. This varies highly from one setup to another. Here are a few basic metrics to monitor.
* CPU Usage
* Memory Usage
* Disk Usage (I/O)
* Disk Space
* Bandwidth

You can run the node and quickly establish baselines for some of these like disk i/o, bandwidth etc.


## Validator and Supporting Processes

### Prometheus Metrics

Axelar runs using Tendermint consensus. Several metrics are exposed via the prometheus metrics endpoint by Tendermint. You can find details on the metrics available through tendermint [here](https://docs.tendermint.com/v0.34/tendermint-core/metrics.html). All these metrics are prefixed with `tendermint` e.g `consensus_height` will be `tendermint_consensus_height`.


Here are a few prometheus metrics to alert on, we have also provided sample queries for prometheus where helpful.

#### Consensus Height

Make sure the that chain is making progress

**Metric Key**: tendermint_consensus_height

**Sample Query**:
```
rate(tendermint_consensus_height[1m])*60 < 5
```

#### P2P Peers

Make sure that there are enough peers you are connected to.

<Callout type="warning" emoji="⚠️">
  When Running a validator behind sentry nodes, the number of peers for the validator will be the number of sentry nodes you have deployed as opposed to the entire network. In this case you want to alert only if the validators peers are less than the number of sentry nodes.
</Callout>

**Metric Key**: tendermint_p2p_peers

**Sample Query**:
```
tendermint_p2p_peers < 10
```

### Other Checks

Other than metrics provided by prometheus itself, there are also some checks you can run periodically to determine the validator is healthy.

#### Restarts

This check will differ from one setup to another, but it is worth keeping an eye on any of the three processes (axelard, vald and tofnd) restarting.

#### Health Check

The health check command should be run to ensure that validator is healthy. The command to run is:
```sh
$AXELARD_HOME/bin/axelard health-check --tofnd-host localhost --operator-addr {VALOPER_ADDR}
```

This should print out an output like this:
```sh
tofnd check: passed
broadcaster check: passed
operator check: passed
```

You can create a script that can check output of this command. For instance to ensure tofnd status is passed you can use something like:
```sh
$AXELARD_HOME/bin/axelard health-check --tofnd-host localhost --operator-addr {VALOPER_ADDR} | grep tofnd | awk -F: '{print $2}' | tr -d ' '
```

#### Chain Maintainers

For all the EVM chains that are active on the network, make sure that your validator is a chain maintainer. To retreieve all chains run the following query:
```sh
axelard q evm chains
```

The query for a specific chain is:
```sh
axelard q nexus chain-maintainers {CHAIN_NAME}
```

To test if your validator is a chain-maintainer, retrieve the `axelarvaloper` address and look for that in the output. For example, for ethereum, the command would look like:
```sh
axelard q nexus chain-maintainers ethereum | grep {VALOPER_ADDR}
```

#### Logs

Monitor the logs for panics. The exact alert will depend on how you store logs.
