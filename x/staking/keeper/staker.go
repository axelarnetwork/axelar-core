package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/staking/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type staker struct {
	totalPower        func(ctx sdk.Context) (power sdk.Int)
	validator         func(ctx sdk.Context, address sdk.ValAddress) (exported.Validator, error)
	iterateValidators func(ctx sdk.Context, fn func(index int64, validator exported.Validator) (stop bool))
	latestSnapshot    func(ctx sdk.Context) (exported.Snapshot, error)
}

func NewStaker(keeper Keeper) exported.Staker {

	return &staker{
		totalPower:        keeper.GetLastTotalPower,
		validator:         keeper.Validator,
		iterateValidators: keeper.IterateValidators,
		latestSnapshot:    keeper.GetLatestSnapshot,
	}
}

func (s *staker) GetLastTotalPower(ctx sdk.Context) (power sdk.Int) {

	return s.totalPower(ctx)
}

func (s *staker) Validator(ctx sdk.Context, address sdk.ValAddress) (exported.Validator, error) {

	return s.validator(ctx, address)

}

func (s *staker) IterateValidators(ctx sdk.Context, fn func(index int64, validator exported.Validator) (stop bool)) {

	s.iterateValidators(ctx, fn)

}

func (s *staker) GetLatestSnapshot(ctx sdk.Context) (exported.Snapshot, error) {

	return s.latestSnapshot(ctx)

}
