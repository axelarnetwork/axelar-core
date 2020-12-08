package types

import (
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	ssExported "github.com/axelarnetwork/axelar-core/x/snapshotting/exported"
	vExported "github.com/axelarnetwork/axelar-core/x/voting/exported"
)

type Broadcaster interface {
	bcExported.Broadcaster
}

type Snapshotter interface {
	ssExported.Snapshotter
}

type Voter interface {
	vExported.Voter
}
