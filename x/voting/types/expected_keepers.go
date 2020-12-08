package types

import (
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshotting "github.com/axelarnetwork/axelar-core/x/snapshotting/exported"
)

type Broadcaster interface {
	broadcast.Broadcaster
}

type Snapshotter interface {
	snapshotting.Snapshotter
}
