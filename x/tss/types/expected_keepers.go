package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . StakingKeeper Snapshotter Nexus MultiSigKeeper

// Snapshotter provides access to the snapshot functionality
type Snapshotter interface {
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

// Nexus provides access to the nexus functionality
type Nexus interface {
	GetChains(ctx sdk.Context) []nexus.Chain
}

// StakingKeeper adopts the methods from "github.com/cosmos/cosmos-sdk/x/staking/exported" that are
// actually used by this module
type StakingKeeper interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
}

// MultiSigKeeper provides multisig functionality
type MultiSigKeeper interface {
	GetKey(ctx sdk.Context, keyID multisig.KeyID) (multisig.Key, bool)
	GetActiveKeyIDs(ctx sdk.Context, chainName nexus.ChainName) []multisig.KeyID
}
