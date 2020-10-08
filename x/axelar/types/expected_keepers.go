package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BridgeKeeper interface {
	TrackAddress(ctx sdk.Context, address string) error
}

type TSSKeeper interface {
}
