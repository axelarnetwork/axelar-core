package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

type Broadcaster interface {
	broadcast.Broadcaster
}

type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
}
