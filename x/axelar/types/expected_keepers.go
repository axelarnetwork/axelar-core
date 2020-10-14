package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

type BridgeKeeper interface {
	TrackAddress(ctx sdk.Context, address string) error
	VerifyTx(ctx sdk.Context, tx exported.ExternalTx) bool
}
