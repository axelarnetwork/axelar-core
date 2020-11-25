package mock

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

var _ stExported.Staker = TestStaker{}

type TestStaker struct {
	validators map[string]stExported.Validator
	totalPower int64
}

func (s TestStaker) GetLatestSnapshot(_ sdk.Context) (stExported.Snapshot, error) {
	panic("implement me")
}

func (s TestStaker) IterateValidators(_ sdk.Context, fn func(index int64, validator stExported.Validator) (stop bool)) {
	var i int64 = 0
	for _, v := range s.validators {
		stop := fn(i, v)
		if stop {
			break
		}
		i++
	}
}

func NewTestStaker(validators ...stExported.Validator) TestStaker {
	staker := TestStaker{map[string]stExported.Validator{}, 0}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
		staker.totalPower += val.Power
	}
	return staker
}

func (s TestStaker) GetLastTotalPower(_ sdk.Context) (power sdk.Int) {
	return sdk.NewInt(s.totalPower)
}

func (s TestStaker) Validator(_ sdk.Context, address sdk.ValAddress) (stExported.Validator, error) {
	v, ok := s.validators[address.String()]
	if !ok {
		return stExported.Validator{}, fmt.Errorf("validator does not exist")
	}
	return v, nil
}

func (s TestStaker) GetAllValidators(_ sdk.Context) []stExported.Validator {
	var vals []stExported.Validator
	for _, v := range s.validators {
		vals = append(vals, v)
	}
	return vals
}
