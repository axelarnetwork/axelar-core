package bridge

import sdk "github.com/cosmos/cosmos-sdk/types"

type Keeper interface {
	TrackAddress(ctx sdk.Context, address string) error
}
