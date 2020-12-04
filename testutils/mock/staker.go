package mock

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	staking "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

var _ staking.Staker = Staker{}

type Staker struct {
	snapshots  []staking.Snapshot
	validators map[string]staking.Validator
}

func NewTestStaker(blockHeight int64, validators ...staking.Validator) Staker {
	staker := Staker{snapshots: []staking.Snapshot{snapshot(blockHeight, validators)}, validators: map[string]staking.Validator{}}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
	}
	return staker
}

func (s Staker) Snapshot(blockHeight int64, validators ...staking.Validator) {
	for _, val := range validators {
		s.validators[val.Address.String()] = val
	}

	s.snapshots = append(s.snapshots, snapshot(blockHeight, validators))
}

func snapshot(blockHeight int64, validators []staking.Validator) staking.Snapshot {
	var totalPower int64
	for _, val := range validators {
		totalPower += val.Power
	}
	return staking.Snapshot{
		Validators: validators,
		Timestamp:  time.Now(),
		Height:     blockHeight,
		TotalPower: sdk.NewInt(totalPower),
	}
}

func (s Staker) Validator(_ sdk.Context, address sdk.ValAddress) (staking.Validator, bool) {
	v, ok := s.validators[address.String()]
	if !ok {
		return staking.Validator{}, false
	}
	return v, true
}

func (s Staker) GetAllValidators() []staking.Validator {
	var vals []staking.Validator
	for _, v := range s.validators {
		vals = append(vals, v)
	}
	return vals
}

func (s Staker) GetLatestRound(_ sdk.Context) int64 {
	return int64(len(s.snapshots) - 1)
}

func (s Staker) GetSnapshot(_ sdk.Context, round int64) (staking.Snapshot, bool) {
	if round >= int64(len(s.snapshots)) {
		return staking.Snapshot{}, false
	}
	return s.snapshots[round], true
}

func (s Staker) GetLatestSnapshot(ctx sdk.Context) (staking.Snapshot, bool) {
	return s.GetSnapshot(ctx, int64(len(s.snapshots)-1))
}
