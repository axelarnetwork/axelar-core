package keeper

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	snapshotMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"
)

var encCfg appParams.EncodingConfig

// Cases to test
var testCases = []struct {
	numValidators, totalPower int
}{
	{
		numValidators: 5,
		totalPower:    50,
	},
	{
		numValidators: 10,
		totalPower:    100,
	},
	{
		numValidators: 3,
		totalPower:    10,
	},
}

func init() {
	// Necessary if tests execute with the real sdk staking keeper
	encCfg = testutils.MakeEncodingConfig()
	encCfg.Amino.RegisterConcrete("", "string", nil)
	staking.RegisterLegacyAminoCodec(encCfg.Amino)

}

func TestTakeSnapshot_WithSubsetSize(t *testing.T) {
	subsetSize := int64(3)
	validators := genValidators(t, 5, 500)
	staker := newMockStaker(validators...)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
	slashingKeeper := &snapshotMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (exported.ValidatorInfo, bool) {
			newInfo := slashingtypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			retinfo := exported.ValidatorInfo{ValidatorSigningInfo: newInfo}
			return retinfo, true
		},
	}
	broadcasterMock := &snapshotMock.BroadcasterMock{
		GetProxyFunc: func(_ sdk.Context, principal sdk.ValAddress) sdk.AccAddress {
			for _, v := range validators {
				if bytes.Equal(principal.Bytes(), v.GetOperator()) {
					return rand.Bytes(sdk.AddrLen)
				}
			}
			return nil
		},
	}
	tssMock := &snapshotMock.TssMock{
		GetValidatorDeregisteredBlockHeightFunc: func(ctx sdk.Context, valAddr sdk.ValAddress) int64 {
			return 0
		},
		GetMinBondFractionPerShareFunc: func(sdk.Context) utils.Threshold {
			return utils.Threshold{Numerator: 1, Denominator: 200}
		},
		GetTssSuspendedUntilFunc: func(sdk.Context, sdk.ValAddress) int64 { return 0 },
	}

	keeper := NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("staking"), snapSubspace, broadcasterMock, staker, slashingKeeper, tssMock)
	keeper.SetParams(ctx, types.DefaultParams())

	_, _, err := keeper.TakeSnapshot(ctx, subsetSize, tss.WeightedByStake)
	assert.NoError(t, err)

	actual, ok := keeper.GetSnapshot(ctx, 0)
	assert.True(t, ok)
	assert.Equal(t, int(subsetSize), len(actual.Validators))
}

// Tests the snapshot functionality
func TestSnapshots(t *testing.T) {
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test-%d", i), func(t *testing.T) {
			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
			validators := genValidators(t, testCase.numValidators, testCase.totalPower)
			staker := newMockStaker(validators...)

			assert.True(t, staker.GetLastTotalPower(ctx).Equal(sdk.NewInt(int64(testCase.totalPower))))

			snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")

			slashingKeeper := &snapshotMock.SlasherMock{
				GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (exported.ValidatorInfo, bool) {
					newInfo := slashingtypes.NewValidatorSigningInfo(
						address,
						int64(0),        // height at which validator was first a candidate OR was unjailed
						int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
						time.Unix(0, 0), // jailed until
						false,           // tomstoned
						int64(0),        // missed blocks
					)
					retinfo := exported.ValidatorInfo{ValidatorSigningInfo: newInfo}
					return retinfo, true
				},
			}

			broadcasterMock := &snapshotMock.BroadcasterMock{
				GetProxyFunc: func(_ sdk.Context, principal sdk.ValAddress) sdk.AccAddress {
					for _, v := range validators {
						if bytes.Equal(principal.Bytes(), v.GetOperator()) {
							return rand.Bytes(sdk.AddrLen)
						}
					}
					return nil
				},
			}

			tssMock := &snapshotMock.TssMock{
				GetValidatorDeregisteredBlockHeightFunc: func(ctx sdk.Context, valAddr sdk.ValAddress) int64 {
					return 0
				},
				GetMinBondFractionPerShareFunc: func(sdk.Context) utils.Threshold {
					return utils.Threshold{Numerator: 1, Denominator: 200}
				},
				GetTssSuspendedUntilFunc: func(sdk.Context, sdk.ValAddress) int64 { return 0 },
			}

			keeper := NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("staking"), snapSubspace, broadcasterMock, staker, slashingKeeper, tssMock)
			keeper.SetParams(ctx, types.DefaultParams())

			_, ok := keeper.GetSnapshot(ctx, 0)

			assert.False(t, ok)
			assert.Equal(t, int64(-1), keeper.GetLatestCounter(ctx))

			_, ok = keeper.GetLatestSnapshot(ctx)

			assert.False(t, ok)

			_, _, err := keeper.TakeSnapshot(ctx, 0, tss.WeightedByStake)

			assert.NoError(t, err)

			snapshot, ok := keeper.GetSnapshot(ctx, 0)

			assert.True(t, ok)
			assert.Equal(t, int64(0), keeper.GetLatestCounter(ctx))
			for i, val := range validators {
				assert.Equal(t, val.GetConsensusPower(), snapshot.Validators[i].GetConsensusPower())
				assert.Equal(t, val.GetOperator(), snapshot.Validators[i].GetOperator())
			}

			_, _, err = keeper.TakeSnapshot(ctx, 0, tss.WeightedByStake)

			assert.Error(t, err)

			ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultParams().LockingPeriod + 100))

			_, _, err = keeper.TakeSnapshot(ctx, 0, tss.WeightedByStake)

			assert.NoError(t, err)

			snapshot, ok = keeper.GetSnapshot(ctx, 1)

			assert.True(t, ok)
			assert.Equal(t, keeper.GetLatestCounter(ctx), int64(1))
			for i, val := range validators {
				assert.Equal(t, val.GetConsensusPower(), snapshot.Validators[i].GetConsensusPower())
				assert.Equal(t, val.GetOperator(), snapshot.Validators[i].GetOperator())
			}
		})
	}
}

// This function returns a set of validators whose voting power adds up to the specified total power
func genValidators(t *testing.T, numValidators, totalConsPower int) []stakingtypes.ValidatorI {
	t.Logf("Total Power: %v", totalConsPower)

	validators := make([]stakingtypes.ValidatorI, numValidators)

	quotient, remainder := totalConsPower/numValidators, totalConsPower%numValidators

	for i := 0; i < numValidators; i++ {
		power := quotient
		if i == 0 {
			power += remainder
		}

		protoPK, err := cryptocodec.FromTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
		if err != nil {
			panic(err)
		}

		pk, err := codectypes.NewAnyWithValue(protoPK)
		if err != nil {
			panic(err)
		}

		validators[i] = staking.Validator{
			OperatorAddress: sdk.ValAddress(rand.Bytes(sdk.AddrLen)).String(),
			Tokens:          sdk.TokensFromConsensusPower(int64(power)),
			Status:          stakingtypes.Bonded,
			ConsensusPubkey: pk,
		}
	}

	return validators
}

var _ types.StakingKeeper = mockStaker{}

type mockStaker struct {
	validators []stakingtypes.ValidatorI
	totalPower sdk.Int
}

func newMockStaker(validators ...stakingtypes.ValidatorI) *mockStaker {
	keeper := &mockStaker{
		make([]stakingtypes.ValidatorI, 0),
		sdk.ZeroInt(),
	}

	for _, val := range validators {
		keeper.validators = append(keeper.validators, val)
		keeper.totalPower = keeper.totalPower.AddRaw(val.GetConsensusPower())
	}

	return keeper
}

func (k mockStaker) GetLastTotalPower(_ sdk.Context) (power sdk.Int) {
	return k.totalPower
}

func (k mockStaker) IterateBondedValidatorsByPower(_ sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	for i, val := range k.validators {
		if fn(int64(i), val) {
			return
		}
	}
}

func (k mockStaker) Validator(_ sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
	for _, validator := range k.validators {
		if bytes.Equal(validator.GetOperator(), addr) {
			return validator
		}
	}

	return nil
}
