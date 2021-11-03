package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	tofnd2 "github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TofndClient TofndKeyGenClient TofndSignClient Voter StakingKeeper TSSKeeper Snapshotter Nexus Rewarder

// Snapshotter provides snapshot functionality
type Snapshotter = snapshot.Snapshotter

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain string) (nexus.Chain, bool)
	GetChains(ctx sdk.Context) []nexus.Chain
}

// Voter provides voting functionality
type Voter interface {
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
	GetPoll(ctx sdk.Context, pollKey vote.PollKey) vote.Poll
}

// InitPoller is a minimal interface to start a poll
type InitPoller = interface {
	// Deprecated: InitializePollWithSnapshot will be removed soon
	InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
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
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
}

// TSSKeeper provides keygen and signing functionality
type TSSKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetParams(ctx sdk.Context, p Params)
	GetParams(ctx sdk.Context) (params Params)
	GetRouter() Router
	SetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID, recoveryInfo []byte)
	HasPrivateRecoveryInfos(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) bool
	GetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) []byte
	SetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID, recoveryInfo []byte)
	GetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) []byte
	DeleteAllRecoveryInfos(ctx sdk.Context, keyID exported.KeyID)
	GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole, keyType exported.KeyType) (exported.KeyRequirement, bool)
	GetTssSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus)
	SetSig(ctx sdk.Context, sigID string, signature []byte)
	GetKeyForSigID(ctx sdk.Context, sigID string) (exported.Key, bool)
	DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool
	PenalizeCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd2.MessageOut_CriminalList_Criminal_CrimeType)
	StartSign(ctx sdk.Context, info exported.SignInfo, snapshotter Snapshotter, voter InitPoller) error
	StartKeygen(ctx sdk.Context, voter Voter, keyInfo KeyInfo, snapshot snapshot.Snapshot) error
	SetAvailableOperator(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...exported.KeyID)
	GetAvailableOperators(ctx sdk.Context, keyIDs ...exported.KeyID) []sdk.ValAddress
	GetKey(ctx sdk.Context, keyID exported.KeyID) (exported.Key, bool)
	SetKey(ctx sdk.Context, keyID exported.KeyID, key ecdsa.PublicKey)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID exported.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool)
	DoesValidatorParticipateInKeygen(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) bool
	DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID)
	DeleteInfoForSig(ctx sdk.Context, sigID string)
	DeleteParticipantsInKeygen(ctx sdk.Context, keyID exported.KeyID)
	DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID)
	SetSigStatus(ctx sdk.Context, sigID string, status exported.SigStatus)
	GetSignParticipants(ctx sdk.Context, sigID string) []string
	SelectSignParticipants(ctx sdk.Context, snapshotter Snapshotter, info exported.SignInfo, snap snapshot.Snapshot) ([]snapshot.Validator, []snapshot.Validator, error)
	GetSignParticipantsAsJSON(ctx sdk.Context, sigID string) []byte
	GetSignParticipantsSharesAsJSON(ctx sdk.Context, sigID string) []byte
	SetInfoForSig(ctx sdk.Context, sigID string, info exported.SignInfo)
	GetInfoForSig(ctx sdk.Context, sigID string) (exported.SignInfo, bool)
	AssertMatchesRequirements(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error
	GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool)
	SetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain, keyIDs []exported.KeyID)
	SetKeyInfo(ctx sdk.Context, info KeyInfo)
	GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold
	GetHeartbeatPeriodInBlocks(ctx sdk.Context) int64
	GetOldActiveKeys(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) ([]exported.Key, error)
	GetMaxSimultaneousSignShares(ctx sdk.Context) int64

	SubmitPubKeys(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress, pubKeys ...[]byte) bool
	GetMultisigPubKeyCount(ctx sdk.Context, keyID exported.KeyID) int64
	IsMultisigKeygenCompleted(ctx sdk.Context, keyID exported.KeyID) bool
	GetKeyType(ctx sdk.Context, keyID exported.KeyID) exported.KeyType
	GetMultisigPubKeyTimeout(ctx sdk.Context, keyID exported.KeyID) (int64, bool)
	HasValidatorSubmittedMultisigPubKey(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress) bool
	GetParticipantsInKeygen(ctx sdk.Context, keyID exported.KeyID) []sdk.ValAddress
	DeleteMultisigKeygen(ctx sdk.Context, keyID exported.KeyID)
}

// Rewarder provides reward functionality
type Rewarder interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}
