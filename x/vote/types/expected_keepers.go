package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
