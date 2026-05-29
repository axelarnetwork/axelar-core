package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:generate moq -pkg mock -out ./mock/types.go . RewardPool

type Reward struct {
	Validator sdk.ValAddress
	Coin      sdk.Coin
}

// RewardPool represents a pool of rewards
type RewardPool interface {
	AddReward(sdk.ValAddress, sdk.Coin)
	AddRewards([]Reward)
	ClearRewards(sdk.ValAddress)
	ReleaseRewards(sdk.ValAddress) error
}

// Refundable interface is used to register refundable message
type Refundable interface {
	sdk.Msg
}
