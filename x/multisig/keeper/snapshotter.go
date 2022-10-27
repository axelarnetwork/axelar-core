package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -pkg mock -out ./mock/snapshotter.go . Snapshotter

// Snapshotter is an interface to create snapshots for multisig keygen
type Snapshotter interface {
	CreateSnapshot(ctx sdk.Context, threshold utils.Threshold) (snapshot.Snapshot, error)
	GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress
}

var _ Snapshotter = SnapshotCreator{}

// SnapshotCreator is an implementation of Snapshotter
type SnapshotCreator struct {
	keygen      types.KeygenParticipator
	snapshotter types.Snapshotter
	staker      types.Staker
	slasher     types.Slasher
}

// NewSnapshotCreator is the constructor for snapshot creator
func NewSnapshotCreator(keygen types.KeygenParticipator, snapshotter types.Snapshotter, staker types.Staker, slasher types.Slasher) SnapshotCreator {
	return SnapshotCreator{
		keygen:      keygen,
		snapshotter: snapshotter,
		staker:      staker,
		slasher:     slasher,
	}
}

// GetOperator returns the operator of the given proxy
func (sc SnapshotCreator) GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	return sc.snapshotter.GetOperator(ctx, proxy)
}

// CreateSnapshot creates a snapshot for multisig keygen
func (sc SnapshotCreator) CreateSnapshot(ctx sdk.Context, threshold utils.Threshold) (snapshot.Snapshot, error) {
	isTombstoned := func(v snapshot.ValidatorI) bool {
		consAdd, err := v.GetConsAddr()
		if err != nil {
			return true
		}

		return sc.slasher.IsTombstoned(ctx, consAdd)
	}

	isProxyActive := func(v snapshot.ValidatorI) bool {
		proxy, isActive := sc.snapshotter.GetProxy(ctx, v.GetOperator())

		return isActive && !sc.keygen.HasOptedOut(ctx, proxy)
	}

	filter := funcs.And(
		funcs.Not(snapshot.ValidatorI.IsJailed),
		funcs.Not(isTombstoned),
		isProxyActive,
	)

	candidates := slices.Map(sc.staker.GetBondedValidatorsByPower(ctx), stakingTypes.Validator.GetOperator)
	return sc.snapshotter.CreateSnapshot(ctx, candidates, filter, snapshot.QuadraticWeightFunc, threshold)
}
