package keeper

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/slices"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

//go:generate moq -pkg mock -out ./mock/snapshotter.go . Snapshotter

type Snapshotter interface {
	CreateSnapshot(ctx sdk.Context, threshold utils.Threshold) (snapshot.Snapshot, error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

var _ Snapshotter = SnapshotCreator{}

type SnapshotCreator struct {
	snapshotter types.Snapshotter
	staker      types.Staker
	slasher     types.Slasher
}

func (sc SnapshotCreator) GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	return sc.snapshotter.GetOperator(ctx, proxy)
}

func (sc SnapshotCreator) CreateSnapshot(ctx sdk.Context, threshold utils.Threshold) (snapshot.Snapshot, error) {
	filter := func(v snapshot.ValidatorI) bool {
		if v.IsJailed() {
			return false
		}

		consAdd, err := v.GetConsAddr()
		if err != nil {
			return false
		}

		if sc.slasher.IsTombstoned(ctx, consAdd) {
			return false
		}

		_, isActive := sc.snapshotter.GetProxy(ctx, v.GetOperator())
		return isActive
	}

	candidates := slices.Map(sc.staker.GetBondedValidatorsByPower(ctx), stakingTypes.Validator.GetOperator)
	return sc.snapshotter.CreateSnapshot(ctx, candidates, filter, snapshot.QuadraticWeightFunc, threshold)
}
