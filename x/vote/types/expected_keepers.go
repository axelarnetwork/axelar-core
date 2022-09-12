package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . Voter Snapshotter StakingKeeper Rewarder

// Voter provides vote keeper functionality
type Voter interface {
	Logger(ctx sdk.Context) log.Logger
	GetVoteRouter() VoteRouter
	GetPoll(ctx sdk.Context, id exported.PollID) (exported.Poll, bool)
	GetPollQueue(ctx sdk.Context) utils.KVQueue
	DeletePoll(ctx sdk.Context, pollID exported.PollID)
	GetParams(ctx sdk.Context) (params Params)
}

// Snapshotter provides snapshot functionality
type Snapshotter interface {
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
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
