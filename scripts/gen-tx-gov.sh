#! /bin/bash

if test "$#" -ne 2
then
   echo "error: 2 parameters are expected (patch version and upgrade height number)"
   echo "example : ./gen-tx-gov.sh v0.16.1 1336350"
    exit
fi

if [[ ! "$1" =~ v[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3} ]]; 
then 
   echo "error: a semantic tag is expected as first parameter"
   echo  "example v0.16.1"
   exit
fi

re='^[0-9]+$'
if ! [[ "$2" =~ $re ]]
then
   echo "error: an integer is expected as second parameter" >&2
   exit
fi

PATCH="$1"
MINOR=$(echo $PATCH | sed 's/..$//g')
HEIGHT_NUMBER="$2"

# LINUX AMD64
curl --silent -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-linux-amd64-${PATCH}.zip.sha256"
LX_AMD64_HASH=$(cat axelard-linux-amd64-${PATCH}.zip.sha256)
rm axelard-linux-amd64-${PATCH}.zip.sha256

# LINUX ARM64
curl --silent -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-linux-arm64-${PATCH}.zip.sha256"
LX_ARM64_HASH=$(cat axelard-linux-arm64-${PATCH}.zip.sha256)
rm axelard-linux-arm64-${PATCH}.zip.sha256

# DARWIN ARM64
curl --silent -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-darwin-arm64-${PATCH}.zip.sha256"
DW_ARM64_HASH=$(cat axelard-darwin-arm64-${PATCH}.zip.sha256)
rm axelard-darwin-arm64-${PATCH}.zip.sha256

# DARWIN AMD64
curl --silent -LJO "https://github.com/axelarnetwork/axelar-core/releases/download/${PATCH}/axelard-darwin-amd64-${PATCH}.zip.sha256"
DW_AMD64_HASH=$(cat axelard-darwin-amd64-${PATCH}.zip.sha256)
rm axelard-darwin-amd64-${PATCH}.zip.sha256

cat <<EOF > txgov.txt | sed -e "s/LX_AMD64_HASH/$LX_AMD64_HASH/; s/LX_ARM64_HASH/$LX_ARM64_HASH/; s/DW_AMD64_HASH/$DW_AMD64_HASH/; s/DW_ARM64_HASH/$DW_ARM64_HASH/"
axelard tx gov submit-proposal software-upgrade "$MINOR" --upgrade-height $HEIGHT_NUMBER --upgrade-info '{"binaries":{"linux/amd64":"https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/$PATCH/axelard-linux-amd64-$PATCH.zip?checksum=sha256:$LX_AMD64_HASH","linux/arm64":"https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/$PATCH/axelard-linux-arm64-$PATCH.zip?checksum=sha256:$LX_ARM64_HASH","darwin/amd64":"https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/$PATCH/axelard-darwin-amd64-$PATCH.zip?checksum=sha256:$DW_AMD64_HASH","darwin/arm64":"https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/$PATCH/axelard-darwin-arm64-$PATCH.zip?checksum=sha256:$DW_ARM64_HASH"}}' --deposit 100000000uaxl --description  "This proposal is intended to upgrade axelar core to $MINOR" --title "Axelar $MINOR Upgrade Proposal" --from validator --gas auto --gas-adjustment 1.2
EOF

cat txgov.txt
rm txgov.txt
