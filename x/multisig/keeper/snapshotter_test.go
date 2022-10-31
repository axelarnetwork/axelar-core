package keeper_test

import (
	"testing"

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
			GetBondedValidatorsByPowerFunc: func(ctx sdk.Context) []stakingtypes.Validator { return bondedValidators },
		}

		jailedAddr     = rand2.ValAddr()
		tombstonedAddr = rand2.ValAddr()
		inactiveAddr   = rand2.ValAddr()
		activeAddr     = rand2.ValAddr()

		jailedVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return true },
			GetConsAddrFunc: func() (sdk.ConsAddress, error) { return sdk.ConsAddress(jailedAddr), nil },
			GetOperatorFunc: func() sdk.ValAddress { return jailedAddr },
		}
		tombstonedVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() (sdk.ConsAddress, error) { return sdk.ConsAddress(tombstonedAddr), nil },
			GetOperatorFunc: func() sdk.ValAddress { return tombstonedAddr },
		}
		inactiveVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() (sdk.ConsAddress, error) { return sdk.ConsAddress(inactiveAddr), nil },
			GetOperatorFunc: func() sdk.ValAddress { return inactiveAddr },
		}
		activeVal = &mock2.ValidatorIMock{
			IsJailedFunc:    func() bool { return false },
			GetConsAddrFunc: func() (sdk.ConsAddress, error) { return sdk.ConsAddress(activeAddr), nil },
			GetOperatorFunc: func() sdk.ValAddress { return activeAddr },
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
			IsTombstonedFunc: func(_ sdk.Context, consAddr sdk.ConsAddress) bool {
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
			weightFunc func(consensusPower sdk.Uint) sdk.Uint,
			threshold utils.Threshold,
		) (snapshot.Snapshot, error) {
			assert.ElementsMatch(t, candidates, slices.Map(bondedValidators, stakingtypes.Validator.GetOperator))

			assert.False(t, filterFunc(jailedVal))
			assert.False(t, filterFunc(tombstonedVal))
			assert.False(t, filterFunc(inactiveVal))
			assert.True(t, filterFunc(activeVal))

			assert.True(t, sdk.NewUint(5).Equal(weightFunc(sdk.NewUint(25))))
			assert.True(t, sdk.NewUint(6).Equal(weightFunc(sdk.NewUint(36))))
			assert.True(t, sdk.NewUint(16).Equal(weightFunc(sdk.NewUint(256))))
			assert.True(t, sdk.NewUint(9).Equal(weightFunc(sdk.NewUint(99))))

			assert.Equal(t, expectedThreshold, threshold)

			return snapshot.Snapshot{}, nil
		}

	_, _ = creator.CreateSnapshot(rand2.Context(fake.NewMultiStore()), expectedThreshold)
}
