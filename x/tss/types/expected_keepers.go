package types

import (
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

type Broadcaster interface {
	broadcast.Broadcaster
}

type Snapshotter interface {
	snapshot.Snapshotter
}

type Voter interface {
	vote.Voter
}
