package mock

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

var _ exported.Snapshotter = Snapshotter{}

type Snapshotter struct {
	snapshots  []exported.Snapshot
	validators map[string]exported.Validator
}

func NewTestStaker(blockHeight int64, validators ...exported.Validator) Snapshotter {
	staker := Snapshotter{snapshots: []exported.Snapshot{snapshot(blockHeight, validators)}, validators: map[string]exported.Validator{}}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
	}
	return staker
}

func (s Snapshotter) Snapshot(blockHeight int64, validators ...exported.Validator) {
	for _, val := range validators {
		s.validators[val.Address.String()] = val
	}

	s.snapshots = append(s.snapshots, snapshot(blockHeight, validators))
}

func snapshot(blockHeight int64, validators []exported.Validator) exported.Snapshot {
	var totalPower int64
	for _, val := range validators {
		totalPower += val.Power
	}
	return exported.Snapshot{
		Validators: validators,
		Timestamp:  time.Now(),
		Height:     blockHeight,
		TotalPower: sdk.NewInt(totalPower),
	}
}

func (s Snapshotter) Validator(_ sdk.Context, address sdk.ValAddress) (exported.Validator, bool) {
	v, ok := s.validators[address.String()]
	if !ok {
		return exported.Validator{}, false
	}
	return v, true
}

func (s Snapshotter) GetAllValidators() []exported.Validator {
	var vals []exported.Validator
	for _, v := range s.validators {
		vals = append(vals, v)
	}
	return vals
}

func (s Snapshotter) GetLatestRound(_ sdk.Context) int64 {
	return int64(len(s.snapshots) - 1)
}

func (s Snapshotter) GetSnapshot(_ sdk.Context, round int64) (exported.Snapshot, bool) {
	if round >= int64(len(s.snapshots)) {
		return exported.Snapshot{}, false
	}
	return s.snapshots[round], true
}

func (s Snapshotter) GetLatestSnapshot(ctx sdk.Context) (exported.Snapshot, bool) {
	return s.GetSnapshot(ctx, int64(len(s.snapshots)-1))
}
