package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	tofnd2 "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TofndClient TofndKeyGenClient TofndSignClient Voter StakingKeeper TSSKeeper Snapshotter Nexus

// Snapshotter provides snapshot functionality
type Snapshotter = snapshot.Snapshotter

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	GetChains(ctx sdk.Context) []nexus.Chain
}

// Voter provides voting functionality
type Voter interface {
	InitializePoll(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
	GetPoll(ctx sdk.Context, pollKey vote.PollKey) vote.Poll
}

// InitPoller is a minimal interface to start a poll
type InitPoller = interface {
	InitializePoll(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
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
	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator types.ValidatorI) (stop bool))
}

// TSSKeeper provides keygen and signing functionality
type TSSKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) (params Params)
	SetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID string, recoveryInfo []byte)
	HasPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID string) bool
	GetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID string) []byte
	DeleteAllRecoveryInfos(ctx sdk.Context, keyID string)
	GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole) (exported.KeyRequirement, bool)
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus)
	SetSig(ctx sdk.Context, sigID string, signature []byte)
	GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool)
	DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool
	PenalizeCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd2.MessageOut_CriminalList_Criminal_CrimeType)
	ScheduleKeygen(ctx sdk.Context, req StartKeygenRequest) (int64, error)
	ScheduleSign(ctx sdk.Context, info exported.SignInfo) (int64, error)
	DeleteScheduledKeygen(ctx sdk.Context, ID string)
	DeleteScheduledSign(ctx sdk.Context, ID string)
	GetAllKeygenRequestsAtCurrentHeight(ctx sdk.Context) []StartKeygenRequest
	StartKeygen(ctx sdk.Context, voter Voter, keyID string, keyRole exported.KeyRole, snapshot snapshot.Snapshot) error
	SetAvailableOperator(ctx sdk.Context, ID string, ackType exported.AckType, validator sdk.ValAddress) error
	GetAvailableOperators(ctx sdk.Context, ID string, ackType exported.AckType, heightLimit int64) []sdk.ValAddress
	DeleteAvailableOperators(ctx sdk.Context, ID string, ackType exported.AckType)
	IsOperatorAvailable(ctx sdk.Context, ID string, ackType exported.AckType, validator sdk.ValAddress) bool
	LinkAvailableOperatorsToSnapshot(ctx sdk.Context, ID string, ackType exported.AckType, counter int64)
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
	HasKeygenStarted(ctx sdk.Context, keyID string) bool
	DeleteKeygenStart(ctx sdk.Context, keyID string)
	DeleteInfoForSig(ctx sdk.Context, sigID string)
	DeleteParticipantsInKeygen(ctx sdk.Context, keyID string)
	DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID string)
	OperatorIsAvailableForCounter(ctx sdk.Context, counter int64, validator sdk.ValAddress) bool
	SetSigStatus(ctx sdk.Context, sigID string, status exported.SigStatus)
	GetSignParticipants(ctx sdk.Context, sigID string) []string
	SelectSignParticipants(ctx sdk.Context, snapshotter Snapshotter, sigID string, validators []snapshot.Validator) (activeShareCount sdk.Int, excluded []snapshot.Validator, err error)
	GetSignParticipantsAsJSON(ctx sdk.Context, sigID string) []byte
	GetSignParticipantsSharesAsJSON(ctx sdk.Context, sigID string) []byte
	SetInfoForSig(ctx sdk.Context, sigID string, info exported.SignInfo)
	GetInfoForSig(ctx sdk.Context, sigID string) (exported.SignInfo, bool)
	AssertMatchesRequirements(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID string, keyRole exported.KeyRole) error
}
