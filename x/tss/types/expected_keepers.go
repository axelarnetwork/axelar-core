package types

import (
	tssd "github.com/axelarnetwork/tssd/pb"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TSSDClient TSSDKeyGenClient TSSDSignClient

type Broadcaster interface {
	broadcast.Broadcaster
}

type Snapshotter interface {
	snapshot.Snapshotter
}

type Voter interface {
	vote.Voter
}

type TSSDClient interface {
	tssd.GG18Client
}

type TSSDKeyGenClient interface {
	tssd.GG18_KeygenClient
}

type TSSDSignClient interface {
	tssd.GG18_SignClient
}
