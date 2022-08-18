package testutils

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
