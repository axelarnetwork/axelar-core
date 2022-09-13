package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Keeper Snapshotter Staker Slasher Rewarder Nexus

// Keeper provides keeper functionality of this module
type Keeper interface {
	Logger(ctx sdk.Context) log.Logger
	GetCurrentKeyID(ctx sdk.Context, chainName nexus.ChainName) (exported.KeyID, bool)
	GetNextKeyID(ctx sdk.Context, chainName nexus.ChainName) (exported.KeyID, bool)
	GetKeygenSession(ctx sdk.Context, id exported.KeyID) (KeygenSession, bool)
	GetKeygenSessionsByExpiry(ctx sdk.Context, expiry int64) []KeygenSession
	GetKey(ctx sdk.Context, keyID exported.KeyID) (exported.Key, bool)
	SetKey(ctx sdk.Context, key Key)
	DeleteKeygenSession(ctx sdk.Context, id exported.KeyID)
	GetSigningSessionsByExpiry(ctx sdk.Context, expiry int64) []SigningSession
	DeleteSigningSession(ctx sdk.Context, id uint64)
	GetSigRouter() SigRouter
}

// Snapshotter provides snapshot keeper functionality
type Snapshotter interface {
	CreateSnapshot(
		ctx sdk.Context,
		candidates []sdk.ValAddress,
		filterFunc func(snapshot.ValidatorI) bool,
		weightFunc func(consensusPower sdk.Uint) sdk.Uint,
		threshold utils.Threshold,
	) (snapshot.Snapshot, error)
	GetProxy(ctx sdk.Context, operator sdk.ValAddress) (addr sdk.AccAddress, active bool)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Staker provides staking keeper functionality
type Staker interface {
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingTypes.Validator
}

// Slasher provides slashing keeper functionality
type Slasher interface {
	IsTombstoned(ctx sdk.Context, consAddr sdk.ConsAddress) bool
}

// Rewarder provides reward keeper functionality
type Rewarder interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}

// Nexus provides nexus keeper functionality
type Nexus interface {
	GetChain(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool)
	GetChains(ctx sdk.Context) []nexus.Chain
}
