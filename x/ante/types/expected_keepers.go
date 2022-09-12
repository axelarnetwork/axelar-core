package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Permission Staking

// MultiSig provides access to the multisig functionality
type MultiSig interface {
	GetNextKeyID(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool)
	GetActiveKeyIDs(ctx sdk.Context, chain nexus.ChainName) []multisig.KeyID
	GetKey(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool)
}

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
}

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
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
