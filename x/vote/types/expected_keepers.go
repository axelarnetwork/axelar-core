package types

import (
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

type Broadcaster interface {
	broadcast.Broadcaster
}

type Snapshotter interface {
	snapshot.Snapshotter
}
