package testutils

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v4/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
)

// RandomIBCTransfer creates a new IBC transfer
func RandomIBCTransfer() types.IBCTransfer {
	channel := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	transfer := types.NewIBCTransfer(
		rand.AccAddr(),
		rand.AccAddr().String(),
		sdk.NewCoin(rand.Denom(5, 20), sdk.NewInt(rand.PosI64())),
		ibctransfertypes.PortID,
		channel,
		nexus.TransferID(rand.PosI64()),
	)

	return transfer
}

// ClientState creates a new client state
func ClientState() *ibctmtypes.ClientState {
	return ibctmtypes.NewClientState(
		"07-tendermint-0",
		ibctmtypes.DefaultTrustLevel,
		time.Hour*24*7*2,
		time.Hour*24*7*3,
		time.Second*10,
		clienttypes.NewHeight(0, 5),
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
}

// RandomIBCDenom creates an ICS20 token denom ibc/{hash}
func RandomIBCDenom() string {
	return fmt.Sprintf("ibc/%s", rand.HexStr(64))
}

// RandomIBCPath creates an IBC path
func RandomIBCPath() string {
	port := ibctransfertypes.PortID
	return types.NewIBCPath(port, RandomChannel())
}

// RandomCosmosChain creates a types.CosmosChain
func RandomCosmosChain() types.CosmosChain {
	return types.CosmosChain{
		Name:       nexustestutils.RandomChainName(),
		IBCPath:    RandomIBCPath(),
		Assets:     nil,
		AddrPrefix: rand.StrBetween(1, 10),
	}
}

// RandomPacket creates a random ICS-20 packet
func RandomPacket(data ibctransfertypes.FungibleTokenPacketData, sourcePort, sourceChannel, destinationPort, destinationChannel string) ibcchanneltypes.Packet {
	return ibcchanneltypes.NewPacket(
		ibctransfertypes.ModuleCdc.MustMarshalJSON(&data),
		uint64(rand.PosI64()),
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		ibcclienttypes.NewHeight(uint64(rand.PosI64()), uint64(rand.PosI64())),
		uint64(rand.PosI64()),
	)
}

// RandomFullDenom creates a fully qualified IBC denom
func RandomFullDenom() string {
	hops := int(rand.I64Between(0, 1))
	denom := rand.Denom(3, 20)
	for i := 0; i < hops; i++ {
		denom = fmt.Sprintf("%s/%s/%s", rand.StrBetween(1, 10), rand.StrBetween(1, 10), denom)
	}
	return denom
}

// RandomChannel creates an IBC channel
func RandomChannel() string {
	return fmt.Sprintf("%s%d", "channel-", rand.PosI64())
}

// PackPayloadWithVersion prepends the version to the payload
func PackPayloadWithVersion(hexVersion string, payload []byte) []byte {
	return append(funcs.Must(hexutil.Decode(hexVersion))[:], payload...)
}
