package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Tss provides access to the tss functionality
type Tss interface {
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID string) (int64, bool)
	GetKeyUnbondingLockingKeyRotationCount(ctx sdk.Context) int64
	GetRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64
	GetKeyByRotationCount(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, rotationCount int64) (exported.Key, bool)
	GetOldActiveKeys(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) ([]tss.Key, error)
}

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	GetSnapshot(ctx sdk.Context, counter int64) (snapshot.Snapshot, bool)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
	GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool)
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error}


// Staking adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type Staking interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
}
