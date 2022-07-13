package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Permission Staking

// Tss provides access to the tss functionality
type Tss interface {
	GetCurrentKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetNextKeyID(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool)
	GetSnapshotCounterForKeyID(ctx sdk.Context, keyID tss.KeyID) (int64, bool)
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
	GetProxy(ctx sdk.Context, operator sdk.ValAddress) (sdk.AccAddress, bool)
}

// Staking adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type Staking interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
}

// Reward provides access to the reward functionality
type Reward interface {
	SetPendingRefund(ctx sdk.Context, req rewardtypes.RefundMsgRequest, refund rewardtypes.Refund) error
}

// Permission provides access to the permission functionality
type Permission interface {
	GetRole(ctx sdk.Context, address sdk.AccAddress) permission.Role
}
