# Running a testnet

This creates a test network running five validators.

## Starting the network

Run `make docker` to build the binaries. Then, run `docker-compose -f docker-compose.single.yml up genesis` 
to create the initial validator. Lastly, run `docker-compose -f docker-compose.multiNodes.yml up` to add four additional 
local nodes.

## Elevating all nodes to validators

Run `bash makeValidators.sh node2 node3 node4 node5` 

## Testing 
Enter either docker container with `docker exec -it <node name> bash`. 
From here you can test the scavenge module (+ token transfer capabilities).
By default, there is an account _treasury_ that holds a large amount of coins. 
You can create your own account with `scavengeCLI keys add <your account name>`.
Then you can transfer tokens to that new account (and make it usable on the network) with `scavengeCLI tx scavenge transferTokens <address of your account> <XXXfoo with XXX being the amount> --from treasury`.
You can query the address of your created account with `scavengeCLI keys show validator -a`. 