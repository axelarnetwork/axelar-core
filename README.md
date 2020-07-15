# Running a testnet

This creates a node running from a pregenerated genesis file.

## Starting the first/only node

Run `make docker`, then `docker-compose up`

## Adding a node to an existing network

Modify the `.env` file by uncommenting the first line and putting in the correct address to an existing node. 
Then start the node as before. 