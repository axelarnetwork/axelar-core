package mock

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshotting "github.com/axelarnetwork/axelar-core/x/snapshotting/exported"
)

var _ snapshotting.Snapshotter = Snapshotter{}

type Snapshotter struct {
	snapshots  []snapshotting.Snapshot
	validators map[string]snapshotting.Validator
}

func NewTestStaker(blockHeight int64, validators ...snapshotting.Validator) Snapshotter {
	staker := Snapshotter{snapshots: []snapshotting.Snapshot{snapshot(blockHeight, validators)}, validators: map[string]snapshotting.Validator{}}

	for _, val := range validators {
		staker.validators[val.Address.String()] = val
	}
	return staker
}

func (s Snapshotter) Snapshot(blockHeight int64, validators ...snapshotting.Validator) {
	for _, val := range validators {
		s.validators[val.Address.String()] = val
	}

	s.snapshots = append(s.snapshots, snapshot(blockHeight, validators))
}

func snapshot(blockHeight int64, validators []snapshotting.Validator) snapshotting.Snapshot {
	var totalPower int64
	for _, val := range validators {
		totalPower += val.Power
	}
	return snapshotting.Snapshot{
		Validators: validators,
		Timestamp:  time.Now(),
		Height:     blockHeight,
		TotalPower: sdk.NewInt(totalPower),
	}
}

func (s Snapshotter) Validator(_ sdk.Context, address sdk.ValAddress) (snapshotting.Validator, bool) {
	v, ok := s.validators[address.String()]
	if !ok {
		return snapshotting.Validator{}, false
	}
	return v, true
}

func (s Snapshotter) GetAllValidators() []snapshotting.Validator {
	var vals []snapshotting.Validator
	for _, v := range s.validators {
		vals = append(vals, v)
	}
	return vals
}

func (s Snapshotter) GetLatestRound(_ sdk.Context) int64 {
	return int64(len(s.snapshots) - 1)
}

func (s Snapshotter) GetSnapshot(_ sdk.Context, round int64) (snapshotting.Snapshot, bool) {
	if round >= int64(len(s.snapshots)) {
		return snapshotting.Snapshot{}, false
	}
	return s.snapshots[round], true
}

func (s Snapshotter) GetLatestSnapshot(ctx sdk.Context) (snapshotting.Snapshot, bool) {
	return s.GetSnapshot(ctx, int64(len(s.snapshots)-1))
}
