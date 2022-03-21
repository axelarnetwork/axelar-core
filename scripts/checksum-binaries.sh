#! /bin/bash

PATCH="$1"

if test "$#" -ne 1
then
   echo "error: 1 parameters are expected (patch version)"
   echo "example : ./checksum-binaries.sh v0.16.1"
   exit
fi

if [[ ! "$1" =~ v[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3} ]]; 
then 
   echo "error: a semantic tag is expected as first parameter"
   echo  "example v0.16.1"
   exit
fi

for OS in "linux-amd64" "linux-arm64" "darwin-arm64" "darwin-amd64"
do
   curl --silent  -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-${OS}-${PATCH}.zip"
   curl --silent  -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-${OS}-${PATCH}"
   curl --silent  -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-${OS}-${PATCH}.zip.sha256"
   curl --silent  -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-${OS}-${PATCH}.sha256"
   OS_ZIPHASH=$(cat "axelard-${OS}-${PATCH}.zip.sha256")
   OS_ZIPHASH_CALCULATED=$(shasum -a 256 axelard-${OS}-${PATCH}.zip | awk '{print $1}')
   OS_HASH=$(cat "axelard-${OS}-${PATCH}.sha256")
   OS_HASH_CALCULATED=$(shasum -a 256 axelard-${OS}-${PATCH} | awk '{print $1}')
   echo "### $OS ###"
   test "$OS_ZIPHASH" == "$OS_ZIPHASH_CALCULATED" && echo "checksum ok for axelard-$OS-$PATCH.zip" || echo "ERROR: checksum mismatch for axelard-$OS-$PATCH.zip"
   test "$OS_HASH" == "$OS_HASH_CALCULATED" && echo "checksum ok for axelard-$OS-$PATCH" || echo "ERROR: checksum mismatch for axelard-$OS-$PATCH" 
done

