#!/bin/bash

docker exec node1 scavengeCLI rest-server --laddr tcp://0.0.0.0:1317&
printf "\n"
printf "\n"
echo "==== creating new scavenge ===="
printf "\n"
docker exec node1 scavengeCLI tx scavenge createScavenge --from treasury 1foo "sol" "descr" -y

sleep 3
printf "\n"
printf "\n"
echo "==== query from inside the container ===="
printf "\n"
docker exec node1 curl http://localhost:1317/scavenge/list

printf "\n"
printf "\n"
echo "==== query from from host ===="
printf "\n"
curl http://localhost:1317/scavenge/list