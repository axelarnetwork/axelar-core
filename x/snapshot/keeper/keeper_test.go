package keeper_test

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

	"github.com/axelarnetwork/axelar-core/app"
	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	snapshotMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
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
	encCfg = app.MakeEncodingConfig()
	encCfg.Amino.RegisterConcrete("", "string", nil)
}

// Tests the snapshot functionality
func TestSnapshots(t *testing.T) {
	keyRequirement := tss.KeyRequirement{
		KeyRole:                    tss.MasterKey,
		MinKeygenThreshold:         utils.Threshold{Numerator: 5, Denominator: 6},
		SafetyThreshold:            utils.Threshold{Numerator: 2, Denominator: 3},
		KeyShareDistributionPolicy: tss.WeightedByStake,
		MaxTotalShareCount:         75,
		KeygenVotingThreshold:      utils.Threshold{Numerator: 5, Denominator: 6},
		SignVotingThreshold:        utils.Threshold{Numerator: 2, Denominator: 3},
		KeygenTimeout:              250,
		SignTimeout:                250,
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test-%d", i), func(t *testing.T) {
			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
			validators := genValidators(t, testCase.numValidators, testCase.totalPower)
			staker := newMockStaker(validators...)
			var counter int64 = 0
			assert.True(t, staker.GetLastTotalPower(ctx).Equal(sdk.NewInt(int64(testCase.totalPower))))

			snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")

			slashingKeeper := &snapshotMock.SlasherMock{
				GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (slashingtypes.ValidatorSigningInfo, bool) {
					newInfo := slashingtypes.NewValidatorSigningInfo(
						address,
						int64(0),        // height at which validator was first a candidate OR was unjailed
						int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
						time.Unix(0, 0), // jailed until
						false,           // tomstoned
						int64(0),        // missed blocks
					)

					return newInfo, true
				},
				SignedBlocksWindowFunc: func(sdk.Context) int64 { return 100 },
			}

			tssMock := &snapshotMock.TssMock{
				GetMaxMissedBlocksPerWindowFunc: func(sdk.Context) utils.Threshold {
					return tsstypes.DefaultParams().MaxMissedBlocksPerWindow
				},
				GetTssSuspendedUntilFunc: func(sdk.Context, sdk.ValAddress) int64 { return 0 },
				IsOperatorAvailableFunc: func(_ sdk.Context, v sdk.ValAddress, keyIDs ...tss.KeyID) bool {
					for _, validator := range validators {
						if validator.GetOperator().String() == v.String() {
							return true
						}
					}
					return false
				},
			}

			snapshotKeeper := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("staking"), snapSubspace, staker, slashingKeeper, tssMock)
			snapshotKeeper.SetParams(ctx, types.DefaultParams())
			for _, v := range validators {
				addr := rand.AccAddr()
				snapshotKeeper.SetProxyReady(ctx, v.GetOperator(), addr)
				_ = snapshotKeeper.RegisterProxy(ctx, v.GetOperator(), addr)
			}

			_, ok := snapshotKeeper.GetSnapshot(ctx, 0)

			assert.False(t, ok)
			assert.Equal(t, int64(-1), snapshotKeeper.GetLatestCounter(ctx))

			_, ok = snapshotKeeper.GetLatestSnapshot(ctx)

			assert.False(t, ok)

			snapshot, err := snapshotKeeper.TakeSnapshot(ctx, keyRequirement)

			assert.NoError(t, err)
			assert.Equal(t, int64(0), snapshotKeeper.GetLatestCounter(ctx))
			for i, val := range validators {
				assert.Equal(t, val.GetConsensusPower(sdk.DefaultPowerReduction), snapshot.Validators[i].GetSDKValidator().GetConsensusPower(sdk.DefaultPowerReduction))
				assert.Equal(t, val.GetOperator(), snapshot.Validators[i].GetSDKValidator().GetOperator())
			}

			_, err = snapshotKeeper.TakeSnapshot(ctx, keyRequirement)
			assert.Error(t, err)

			ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultParams().LockingPeriod + 100))

			counter++
			_, err = snapshotKeeper.TakeSnapshot(ctx, keyRequirement)

			assert.NoError(t, err)

			snapshot, ok = snapshotKeeper.GetSnapshot(ctx, 1)

			assert.True(t, ok)
			assert.Equal(t, snapshotKeeper.GetLatestCounter(ctx), int64(1))
			for i, val := range validators {
				assert.Equal(t, val.GetConsensusPower(sdk.DefaultPowerReduction), snapshot.Validators[i].GetSDKValidator().GetConsensusPower(sdk.DefaultPowerReduction))
				assert.Equal(t, val.GetOperator(), snapshot.Validators[i].GetSDKValidator().GetOperator())
			}
		})
	}
}

func TestKeeper_RegisterProxy(t *testing.T) {
	var (
		ctx              sdk.Context
		snapshotKeeper   keeper.Keeper
		principalAddress sdk.ValAddress
		staker           *mockStaker
		validators       []staking.ValidatorI
	)

	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
		validators = genValidators(t, 10, 100)
		staker = newMockStaker(validators...)
		principalAddress = validators[rand.I64Between(0, 10)].GetOperator()

		snapshotKeeper = keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("staking"), snapSubspace, staker, &snapshotMock.SlasherMock{}, &snapshotMock.TssMock{})
	}
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		expectedProxy := rand.AccAddr()
		snapshotKeeper.SetProxyReady(ctx, principalAddress, expectedProxy)
		err := snapshotKeeper.RegisterProxy(ctx, principalAddress, expectedProxy)

		assert.NoError(t, err)
		proxy, active := snapshotKeeper.GetProxy(ctx, principalAddress)
		assert.True(t, active)
		assert.Equal(t, expectedProxy, proxy)

	}).Repeat(20))

	t.Run("proxy not ready", testutils.Func(func(t *testing.T) {
		setup()

		expectedProxy := rand.AccAddr()
		err := snapshotKeeper.RegisterProxy(ctx, principalAddress, expectedProxy)

		assert.Error(t, err)

	}).Repeat(20))

	t.Run("unknown validator", testutils.Func(func(t *testing.T) {
		setup()

		address := rand.ValAddr()
		proxy := rand.AccAddr()
		snapshotKeeper.SetProxyReady(ctx, address, proxy)
		err := snapshotKeeper.RegisterProxy(ctx, address, proxy)

		assert.Error(t, err)

	}).Repeat(20))
}

func TestKeeper_DeregisterProxy(t *testing.T) {
	var (
		ctx              sdk.Context
		snapshotKeeper   keeper.Keeper
		principalAddress sdk.ValAddress
		expectedProxy    sdk.AccAddress
		staker           *mockStaker
		validators       []staking.ValidatorI
	)

	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		snapSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snap")
		validators = genValidators(t, 10, 100)
		staker = newMockStaker(validators...)
		principalAddress = validators[rand.I64Between(0, 10)].GetOperator()
		expectedProxy = rand.AccAddr()

		snapshotKeeper = keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("staking"), snapSubspace, staker, &snapshotMock.SlasherMock{}, &snapshotMock.TssMock{})
		snapshotKeeper.SetProxyReady(ctx, principalAddress, expectedProxy)
		if err := snapshotKeeper.RegisterProxy(ctx, principalAddress, expectedProxy); err != nil {
			panic(fmt.Sprintf("setup failed for unit test: %v", err))
		}
	}
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := snapshotKeeper.DeactivateProxy(ctx, principalAddress)

		assert.NoError(t, err)
		proxy, active := snapshotKeeper.GetProxy(ctx, principalAddress)
		assert.False(t, active)
		assert.Equal(t, expectedProxy, proxy)

	}).Repeat(20))

	t.Run("unknown validator", testutils.Func(func(t *testing.T) {
		setup()

		address := rand.ValAddr()
		err := snapshotKeeper.DeactivateProxy(ctx, address)

		assert.Error(t, err)

	}).Repeat(20))

	t.Run("no proxy", testutils.Func(func(t *testing.T) {
		setup()

		var address sdk.ValAddress
		for {
			address = validators[rand.I64Between(0, 10)].GetOperator()
			if !bytes.Equal(principalAddress, address) {
				break
			}
		}

		principalAddress = address
		err := snapshotKeeper.DeactivateProxy(ctx, principalAddress)

		assert.Error(t, err)

	}).Repeat(20))
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
			OperatorAddress: rand.ValAddr().String(),
			Tokens:          sdk.TokensFromConsensusPower(int64(power), sdk.DefaultPowerReduction),
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
	staker := &mockStaker{
		make([]stakingtypes.ValidatorI, 0),
		sdk.ZeroInt(),
	}

	for _, val := range validators {
		staker.validators = append(staker.validators, val)
		staker.totalPower = staker.totalPower.AddRaw(val.GetConsensusPower(sdk.DefaultPowerReduction))
	}

	return staker
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

func (k mockStaker) PowerReduction(ctx sdk.Context) sdk.Int {
	return sdk.DefaultPowerReduction
}
