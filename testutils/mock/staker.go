package mock

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

var _ staking.Staker = Staker{}

type Staker struct {
	validators map[string]staking.Validator
	totalPower int64
}

func (s Staker) GetLatestSnapshot(_ sdk.Context) (staking.Snapshot, error) {
	panic("implement me")
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

func NewTestStaker(validators ...staking.Validator) Staker {
	staker := Staker{map[string]staking.Validator{}, 0}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
		staker.totalPower += val.Power
	}
	return staker
}

func (s Staker) GetLastTotalPower(_ sdk.Context) (power sdk.Int) {
	return sdk.NewInt(s.totalPower)
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
