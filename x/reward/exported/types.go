package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardPool represents a pool of rewards
type RewardPool interface {
	AddReward(sdk.ValAddress, sdk.Coin)
	ClearRewards(sdk.ValAddress)
	ReleaseRewards(sdk.ValAddress) error
}
