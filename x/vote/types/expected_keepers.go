package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// Broadcaster provides access to the proxy-to-validator mapping
type Broadcaster interface {
	// GetPrincipal returns the principal address for a given proxy address. Returns nil if not set.
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
