package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Snapshotter StakingKeeper Rewarder

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetSnapshot(sdk.Context, int64) (snapshot.Snapshot, bool)
}

// StakingKeeper provides functionality of the staking module
type StakingKeeper interface {
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	PowerReduction(sdk.Context) sdk.Int
	GetLastTotalPower(sdk.Context) sdk.Int
}

// Rewarder provides reward functionality
type Rewarder interface {
	GetPool(ctx sdk.Context, name string) reward.RewardPool
}
