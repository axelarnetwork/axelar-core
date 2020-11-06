# axelar-core

The axelar-core app based on the Cosmos SDK is the main application of the axelar network.
This repository is used to build the necessary binaries and docker image to run a core validator.
See the axelarnetwork/axelarate repo for instructions to run a node. 

## Building binaries locally

Execute `make build` to create local binaries for the validator node. 
They are created in the `./bin` folder.

## Creating docker images
To create a regular docker image for the validator, execute `make docker-image`.
This creates the image axelar/core:latest.

To create a docker image for debugging (with [delve](https://github.com/go-delve/delve)), execute `make docker-image-debug`.
This creates the image axelar/core-debug:latest.

## Interacting with a local validator node
With a local (dockerized) validator running, the `axelarcli` binary can be used to interact with the node.
Run `./bin/axelarcli --help` after building the binaries to get information about the available commands.

## Installing Prometheus + Grafana
1. Modify axelarate/monitoring/prometheus/prometheus-2.20.1.linux-amd64/prometheus.yml to add
`  - job_name: 'tendermint'
    static_configs:
      - targets: ['docker.for.mac.localhost:26660']
        labels:
          group: 'tendermint'`
 at the end of the file. 
 2. Run `docker run -p 9090:9090 -v <FIXPATH>/axelarate/monitoring/prometheus/prometheus-2.20.1.linux-amd64/file/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus`
 3. Setup a grafana account online.
 4. In my case, `https://axelartest.grafana.net/` is the domain I can register.
 5. Add a data source with URL `http://localhost:9090`, access: `browser`. 
 6. Import a dashboard for Cosmos. One option is `https://github.com/zhangyelong/cosmos-dashboard` dashboard ID: `11036`
 7. In the top right corner set auto-refresh to `5` seconds or something small
 8. Some data does not refresh consistently. As a temp-fix, I do to `edit` on the datafeed, turn on and off `instant` toggle. 
