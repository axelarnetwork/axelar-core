# Running a testnet

This creates a node running from a pregenerated genesis file.

## Starting the first/only node

Run `make docker`, then `docker-compose -f docker-compose-run.yml up`

## Adding a node to an existing network

Modify the `.env` file by uncommenting the first line and putting in the correct address to an existing node. 
Then start the node as before. 

## Testing 
Enter the docker container with `docker exec -it node1 bash`. 
From here you can test the scavenge module (+ token transfer capabilities).
By default, there is an account _treasury_ that holds a large amount of coins. 
You can create your own account with `scavengeCLI keys add <your account name>`.
Then you can transfer tokens to that new account (and make it usable on the network) with `scavengeCLI tx scavenge transferTokens <address of your account> <XXXfoo with XXX being the amount> --from treasury`. 