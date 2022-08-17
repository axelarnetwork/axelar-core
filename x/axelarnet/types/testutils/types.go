package testutils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// RandomIBCTransfer creates a new IBC transfer
func RandomIBCTransfer() types.IBCTransfer {
	channel := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	transfer := types.NewIBCTransfer(rand.AccAddr(), rand.AccAddr().String(), sdk.NewCoin(rand.Denom(5, 20), sdk.NewInt(rand.PosI64())), ibctransfertypes.PortID, channel)
	transfer.Status = types.TransferPending
	funcs.MustNoErr(transfer.SetID(nexus.TransferID(rand.PosI64())))

	return transfer
}
