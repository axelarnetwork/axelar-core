package testutils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// RandomIBCTransfer creates a new IBC transfer
func RandomIBCTransfer() types.IBCTransfer {
	denom := rand.Strings(5, 20).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")).Next()
	channel := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	transfer := types.NewIBCTransfer(rand.AccAddr(), rand.NormalizedStrBetween(5, 20), sdk.NewCoin(denom, sdk.NewInt(rand.PosI64())), ibctransfertypes.PortID, channel)
	transfer.SetID(nexus.TransferID(rand.PosI64()))

	return transfer
}
