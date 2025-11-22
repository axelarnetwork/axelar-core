package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	mock2 "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/utils/slices"
)

func TestSnapshotCreator_CreateSnapshot(t *testing.T) {
	var (
		bondedValidators = slices.Expand(func(int) stakingtypes.Validator {
			return stakingtypes.Validator{OperatorAddress: rand2.ValAddr().String()}
		}, 10)

		staker = &mock.StakerMock{
			GetBondedValidatorsByPowerFunc: func(ctx context.Context) ([]stakingtypes.Validator, error) { return bondedValidators, nil },
		}

		jailedAddr     = rand2.ValAddr()
		tombstonedAddr = rand2.ValAddr()
		inactiveAddr   = rand2.ValAddr()
		activeAddr     = rand2.ValAddr()

		jailedVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return true },
			GetConsAddrFunc: func() ([]byte, error) { return sdk.ConsAddress(jailedAddr), nil },
			GetOperatorFunc: func() string { return jailedAddr.String() },
		}
		tombstonedVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() ([]byte, error) { return sdk.ConsAddress(tombstonedAddr), nil },
			GetOperatorFunc: func() string { return tombstonedAddr.String() },
		}
		inactiveVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() ([]byte, error) { return sdk.ConsAddress(inactiveAddr), nil },
			GetOperatorFunc: func() string { return inactiveAddr.String() },
		}
		activeVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() ([]byte, error) { return sdk.ConsAddress(activeAddr), nil },
			GetOperatorFunc: func() string { return activeAddr.String() },
		}

		snapshotter = &mock.SnapshotterMock{
			GetProxyFunc: func(_ sdk.Context, operator sdk.ValAddress) (sdk.AccAddress, bool) {
				if operator.Equals(inactiveAddr) {
					return nil, false
				} else {
					return rand2.AccAddr(), true
				}
			}}

		slasher = &mock.SlasherMock{
			IsTombstonedFunc: func(_ context.Context, consAddr sdk.ConsAddress) bool {
				return consAddr.Equals(tombstonedAddr)
			}}

		keygen = &mock.KeygenParticipatorMock{HasOptedOutFunc: func(sdk.Context, sdk.AccAddress) bool {
			return false
		}}

		expectedThreshold = testutils.RandThreshold()
	)

	creator := keeper.NewSnapshotCreator(keygen, snapshotter, staker, slasher)

	snapshotter.CreateSnapshotFunc =
		func(
			_ sdk.Context,
			candidates []sdk.ValAddress,
			filterFunc func(snapshot.ValidatorI) bool,
			weightFunc func(consensusPower math.Uint) math.Uint,
			threshold utils.Threshold,
		) (snapshot.Snapshot, error) {
			assert.ElementsMatch(t,
				slices.Map(candidates, func(v sdk.ValAddress) string { return v.String() }),
				slices.Map(bondedValidators, stakingtypes.Validator.GetOperator),
			)

			assert.False(t, filterFunc(jailedVal))
			assert.False(t, filterFunc(tombstonedVal))
			assert.False(t, filterFunc(inactiveVal))
			assert.True(t, filterFunc(activeVal))

			assert.True(t, math.NewUint(5).Equal(weightFunc(math.NewUint(25))))
			assert.True(t, math.NewUint(6).Equal(weightFunc(math.NewUint(36))))
			assert.True(t, math.NewUint(16).Equal(weightFunc(math.NewUint(256))))
			assert.True(t, math.NewUint(9).Equal(weightFunc(math.NewUint(99))))

			assert.Equal(t, expectedThreshold, threshold)

			return snapshot.Snapshot{}, nil
		}

	_, _ = creator.CreateSnapshot(rand2.Context(fake.NewMultiStore(), t), expectedThreshold)
}
