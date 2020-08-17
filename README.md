# Running a testnet

This creates a test network running five validators.

## Starting the network

Run `make docker` to build the binaries. Then, run `docker-compose -f docker-compose.single.yml up genesis` 
to create the initial validator. Lastly, run `docker-compose -f docker-compose.multiNodes.yml up` to add four additional 
local nodes.

## Elevating all nodes to validators

Run `bash makeValidators.sh node2 node3 node4 node5` 

## Creating additional accounts with funds
Enter either docker container with `docker exec -it <node name> bash`. 
From here you can test the scavenge module (+ token transfer capabilities).
By default, there is an account _treasury_ that holds a large amount of coins. 
You can create your own account with `scavengeCLI keys add <your account name>`.
Then you can transfer tokens to that new account (and make it usable on the network) with `scavengeCLI tx scavenge transferTokens <address of your account> <XXXfoo with XXX being the amount> --from treasury`.
You can query the address of your created account with `scavengeCLI keys show validator -a`. 

## Load testing
Enter either docker container with `docker exec -it <node name> bash`. 
Use the command `testCLI tp [txCount] [goroutines] [account_with_funds] [amount] [flags]` for load testing.
Example: `testCLI tp 1000 10 treasury 1foo -y`

**Important: Do not forget the -y flag, otherwise you will have to confirm every single transaction**

## Monitoring with Prometheus + Grafana
To be able to run Prometheus, execute `make prometheus` _once_. This will create a _prometheus_ user 
and a data directory that is owned by that user. As long as you do not delete the directory or the user, 
the command does not have to be re-executed.

Spin up the monitoring node with `docker-compose -f docker-compose.monitor.yml up`. 
It is configured to read Cosmos metrics from the default metrics port of the genesis container (genesis:26660).
Grafana is available at `http://localhost:3000`.

###Configuring Grafana 
The default login is User:`admin` PW:`admin`. When logged in add a new data source at 
Configuration (cog icon) -> Data Sources. Choose Prometheus, change the name to _Cosmos_ and fill in the URL
`http://localhost:9090`. Click _Save & Test_. This should show _Data source is working_. 
To import the predefined dashboard go to Create (plus icon) -> Import. Copy the json from `./grafana/dashboard.json`
into the designated field and click _Load_.

