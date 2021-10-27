package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

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
	IsProxyReady(ctx sdk.Context, operator sdk.ValAddress) bool
}

// Staking adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type Staking interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
}

// Axelarnet provides access to the axelarnet functionality
type Axelarnet interface {
	SetPendingRefund(ctx sdk.Context, req axelarnettypes.RefundMsgRequest, fee sdk.Coin) error
}
