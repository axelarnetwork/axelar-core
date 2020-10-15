package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type BridgeKeeper interface {
	TrackAddress(ctx sdk.Context, address string) error
	VerifyTx(ctx sdk.Context, tx exported.ExternalTx) bool
}

type Broadcaster interface {
	bcExported.Broadcaster
}
