package mock

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

var _ staking.Staker = Staker{}

type Staker struct {
	validators map[string]staking.Validator
	totalPower int64
	round      int64
}

func NewTestStaker(startRound int64, validators ...staking.Validator) Staker {
	staker := Staker{validators: map[string]staking.Validator{}, round: startRound}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
		staker.totalPower += val.Power
	}
	return staker
}

func (s Staker) Validator(_ sdk.Context, address sdk.ValAddress) (staking.Validator, error) {
	v, ok := s.validators[address.String()]
	if !ok {
		return staking.Validator{}, fmt.Errorf("validator does not exist")
	}
	return v, nil
}

func (s Staker) GetAllValidators(_ sdk.Context) []staking.Validator {
	var vals []staking.Validator
	for _, v := range s.validators {
		vals = append(vals, v)
	}
	return vals
}

func (s Staker) GetLatestRound(_ sdk.Context) int64 {
	return s.round
}

func (s Staker) GetSnapshot(_ sdk.Context, round int64) (staking.Snapshot, error) {
	if round != s.round {
		return staking.Snapshot{}, fmt.Errorf("wrong round")
	}
	var vs []staking.Validator
	for _, v := range s.validators {
		vs = append(vs, v)
	}
	return staking.Snapshot{
		Validators: vs,
		Timestamp:  time.Now(),
		Height:     s.round,
		TotalPower: sdk.NewInt(s.totalPower),
	}, nil
}

func (s Staker) GetLatestSnapshot(ctx sdk.Context) (staking.Snapshot, error) {
	return s.GetSnapshot(ctx, s.round)
}

func (s Staker) IterateValidators(_ sdk.Context, fn func(index int64, validator staking.Validator) (stop bool)) {
	var i int64 = 0
	for _, v := range s.validators {
		stop := fn(i, v)
		if stop {
			break
		}
		i++
	}
}
