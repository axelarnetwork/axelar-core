package types

import (
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TSSDClient TSSDKeyGenClient TSSDSignClient

// Broadcaster provides broadcasting functionality
type Broadcaster interface {
	broadcast.Broadcaster
}

// Snapshotter provides validator snapshot functionality
type Snapshotter interface {
	snapshot.Snapshotter
}

// Balancer provides access to the hub functionality
type Balancer interface {
	GetChain(ctx sdk.Context, chain string) (exported.Chain, bool)
}

// Voter provides voting functionality
type Voter interface {
	vote.Voter
}

// TSSDClient wraps around TSSDKeyGenClient and TSSDSignClient
type TSSDClient interface {
	tssd.GG18Client
}

// TSSDKeyGenClient provides keygen functionality
type TSSDKeyGenClient interface {
	tssd.GG18_KeygenClient
}

// TSSDSignClient provides signing functionality
type TSSDSignClient interface {
	tssd.GG18_SignClient
}
