package types

import (
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

// Nexus provides access to the nexus functionality
type Nexus interface {
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

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
}
