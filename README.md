# Running a testnet

This creates a test network running two validators.

## Starting two nodes with one acting as validator

Run `make docker` to build the binaries, then `docker-compose -f docker-compose-multiple.yml up`

## Making the second node a validator

Run `docker exec node2 bash makeValidator.sh node2` 

## Testing 
Enter either docker container with `docker exec -it <node name> bash`. 
From here you can test the scavenge module (+ token transfer capabilities).
By default, there is an account _treasury_ that holds a large amount of coins. 
You can create your own account with `scavengeCLI keys add <your account name>`.
Then you can transfer tokens to that new account (and make it usable on the network) with `scavengeCLI tx scavenge transferTokens <address of your account> <XXXfoo with XXX being the amount> --from treasury`.
You can query the address of your created account with `scavengeCLI keys show validator -a`. 