package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	tofnd2 "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TofndClient TofndKeyGenClient TofndSignClient Voter StakingKeeper

// Broadcaster provides broadcasting functionality
type Broadcaster interface {
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Snapshotter provides validator snapshot functionality
type Snapshotter interface {
	GetLatestSnapshot(ctx sdk.Context) (snapshot.Snapshot, bool)
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy exported.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error)
}

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
}

// Voter provides voting functionality
type Voter interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta, snapshotCounter int64, expireAt int64) error
	DeletePoll(ctx sdk.Context, poll vote.PollMeta)
	TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta vote.PollMeta, data vote.VotingData) error
	Result(ctx sdk.Context, poll vote.PollMeta) vote.VotingData
	GetPoll(ctx sdk.Context, pollMeta vote.PollMeta) *votetypes.Poll
}

// InitPoller is a minimal interface to start a poll
type InitPoller = interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta, snapshotCounter int64, expireAt int64) error
}

// TofndClient wraps around TofndKeyGenClient and TofndSignClient
type TofndClient interface {
	tofnd2.GG20Client
}

// TofndKeyGenClient provides keygen functionality
type TofndKeyGenClient interface {
	tofnd2.GG20_KeygenClient
}

// TofndSignClient provides signing functionality
type TofndSignClient interface {
	tofnd2.GG20_SignClient
}

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) (power sdk.Int)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool)
}
