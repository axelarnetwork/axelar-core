package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -pkg mock -out ./mock/types.go . RewardPool

// RewardPool represents a pool of rewards
type RewardPool interface {
	AddReward(sdk.ValAddress, sdk.Coin)
	ClearRewards(sdk.ValAddress)
	ReleaseRewards(sdk.ValAddress) error
}
