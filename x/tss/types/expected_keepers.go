package types

import (
	"crypto/ecdsa"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . TofndClient TofndKeyGenClient TofndSignClient Voter StakingKeeper TSSKeeper Snapshotter Nexus Rewarder MultiSigKeeper Slasher

// Snapshotter provides snapshot functionality
type Snapshotter = snapshot.Snapshotter

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool)
	GetChains(ctx sdk.Context) []nexus.Chain
}

// Voter provides voting functionality
type Voter interface {
	InitializePoll(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error)
	GetPoll(ctx sdk.Context, pollID vote.PollID) (vote.Poll, bool)
}

// InitPoller is a minimal interface to start a poll
type InitPoller = interface {
	InitializePoll(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error)
}

// TofndClient wraps around TofndKeyGenClient and TofndSignClient
type TofndClient interface {
	tofnd.GG20Client
}

// TofndKeyGenClient provides keygen functionality
type TofndKeyGenClient interface {
	tofnd.GG20_KeygenClient
}

// TofndSignClient provides signing functionality
type TofndSignClient interface {
	tofnd.GG20_SignClient
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
	HasPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) bool
	GetPrivateRecoveryInfo(ctx sdk.Context, sender sdk.ValAddress, keyID exported.KeyID) []byte
	SetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID, recoveryInfo []byte)
	GetGroupRecoveryInfo(ctx sdk.Context, keyID exported.KeyID) []byte
	DeleteKeyRecoveryInfo(ctx sdk.Context, keyID exported.KeyID)
	GetKeyRequirement(ctx sdk.Context, keyRole exported.KeyRole, keyType exported.KeyType) (exported.KeyRequirement, bool)
	GetSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64
	GetSig(ctx sdk.Context, sigID string) (exported.Signature, exported.SigStatus)
	SetSig(ctx sdk.Context, signature exported.Signature)
	DoesValidatorParticipateInSign(ctx sdk.Context, sigID string, validator sdk.ValAddress) bool
	PenalizeCriminal(ctx sdk.Context, criminal sdk.ValAddress, crimeType tofnd.MessageOut_CriminalList_Criminal_CrimeType)
	StartSign(ctx sdk.Context, info exported.SignInfo, snapshotter Snapshotter, voter InitPoller) error
	StartKeygen(ctx sdk.Context, voter Voter, keyInfo KeyInfo, snapshot snapshot.Snapshot) error
	SetAvailableOperator(ctx sdk.Context, validator sdk.ValAddress, keyIDs ...exported.KeyID)
	GetAvailableOperators(ctx sdk.Context, keyIDs ...exported.KeyID) []sdk.ValAddress
	GetKey(ctx sdk.Context, keyID exported.KeyID) (exported.Key, bool)
	SetKey(ctx sdk.Context, key exported.Key)
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool)
	GetCurrentKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.KeyID, bool)
	GetNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) (exported.Key, bool)
	AssignNextKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole, keyID exported.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) error
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID) (int64, bool)
	HasKeygenStarted(ctx sdk.Context, keyID exported.KeyID) bool
	DeleteKeygenStart(ctx sdk.Context, keyID exported.KeyID)
	DeleteInfoForSig(ctx sdk.Context, sigID string)
	DeleteSnapshotCounterForKeyID(ctx sdk.Context, keyID exported.KeyID)
	SetSigStatus(ctx sdk.Context, sigID string, status exported.SigStatus)
	GetSignParticipants(ctx sdk.Context, sigID string) []string
	GetSignParticipantsAsJSON(ctx sdk.Context, sigID string) []byte
	GetSignParticipantsSharesAsJSON(ctx sdk.Context, sigID string) []byte
	SetInfoForSig(ctx sdk.Context, sigID string, info exported.SignInfo)
	GetInfoForSig(ctx sdk.Context, sigID string) (exported.SignInfo, bool)
	AssertMatchesRequirements(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error
	GetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain) ([]exported.KeyID, bool)
	SetExternalKeyIDs(ctx sdk.Context, chain nexus.Chain, keyIDs []exported.KeyID)
	GetExternalMultisigThreshold(ctx sdk.Context) utils.Threshold
	GetHeartbeatPeriodInBlocks(ctx sdk.Context) int64
	GetOldActiveKeys(ctx sdk.Context, chain nexus.Chain, keyRole exported.KeyRole) ([]exported.Key, error)
	GetMaxSimultaneousSignShares(ctx sdk.Context) int64

	SubmitPubKeys(ctx sdk.Context, keyID exported.KeyID, validator sdk.ValAddress, pubKeys ...[]byte) bool
	GetMultisigKeygenInfo(ctx sdk.Context, keyID exported.KeyID) (MultisigKeygenInfo, bool)
	IsMultisigKeygenCompleted(ctx sdk.Context, keyID exported.KeyID) bool
	GetKeyType(ctx sdk.Context, keyID exported.KeyID) exported.KeyType
	DeleteMultisigKeygen(ctx sdk.Context, keyID exported.KeyID)
	GetMultisigPubKeysByValidator(ctx sdk.Context, keyID exported.KeyID, val sdk.ValAddress) ([]ecdsa.PublicKey, bool)
	SubmitSignatures(ctx sdk.Context, sigID string, validator sdk.ValAddress, sigs ...[]byte) bool
	GetMultisigSignInfo(ctx sdk.Context, sigID string) (MultisigSignInfo, bool)
	DeleteMultisigSign(ctx sdk.Context, signID string)
}

// Rewarder provides reward functionality
type Rewarder interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}

// MultiSigKeeper provides multisig functionality
type MultiSigKeeper interface {
	SetKey(ctx sdk.Context, key types.Key)
	AssignKey(ctx sdk.Context, chain nexus.ChainName, id multisig.KeyID) error
	RotateKey(ctx sdk.Context, chain nexus.ChainName) error
	GetKey(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool)
	GetActiveKeyIDs(ctx sdk.Context, chainName nexus.ChainName) []multisig.KeyID
}

// Slasher provides slasher functionality
type Slasher interface {
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (slashingtypes.ValidatorSigningInfo, bool)
	SignedBlocksWindow(ctx sdk.Context) int64
	GetValidatorMissedBlockBitArray(ctx sdk.Context, address sdk.ConsAddress, index int64) bool
}
