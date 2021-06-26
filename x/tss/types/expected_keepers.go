package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	tofnd2 "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TofndClient TofndKeyGenClient TofndSignClient Voter StakingKeeper TSSKeeper Snapshotter Nexus

// Broadcaster provides broadcasting functionality
type Broadcaster interface {
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Snapshotter provides validator snapshot functionality
type Snapshotter interface {
	GetLatestSnapshot(ctx sdk.Context) (snapshot.Snapshot, bool)
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
	TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy exported.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error)
	GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
}

// Voter provides voting functionality
type Voter interface {
	InitPoll(ctx sdk.Context, poll vote.PollMeta, snapshotCounter int64, expireAt int64) error
	DeletePoll(ctx sdk.Context, poll vote.PollMeta)
	TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta vote.PollMeta, data vote.VotingData) (*votetypes.Poll, error)
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
type TSSKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) (params Params)
	SetKeyRequirement(ctx sdk.Context, keyRequirement exported.KeyRequirement)
	GetKeyRequirement(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyRequirement, bool)
	ComputeCorruptionThreshold(ctx sdk.Context, totalShareCount sdk.Int) int64
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	StartSign(ctx sdk.Context, voter InitPoller, keyID string, sigID string, msg []byte, s snapshot.Snapshot) error
	GetSig(ctx sdk.Context, sigID string) (exported.Signature, bool)
	SetSig(ctx sdk.Context, sigID string, signature []byte)
	GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool)
	DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool
	PenalizeSignCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd2.MessageOut_CriminalList_Criminal_CrimeType)
	StartKeygen(ctx sdk.Context, voter Voter, keyID string, snapshot snapshot.Snapshot) error
	GetKey(ctx sdk.Context, keyID string) (exported.Key, bool)
	SetKey(ctx sdk.Context, keyID string, key ecdsa.PublicKey)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (string, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID string) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
	DoesValidatorParticipateInKeygen(ctx sdk.Context, keyID string, validator sdk.ValAddress) bool
	GetMinKeygenThreshold(ctx sdk.Context) utils.Threshold
	GetMinBondFractionPerShare(ctx sdk.Context) utils.Threshold
	DeleteKeygenStart(ctx sdk.Context, keyID string)
	DeleteKeyIDForSig(ctx sdk.Context, sigID string)
	DeleteParticipantsInKeygen(ctx sdk.Context, keyID string)
	DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID string)
}
