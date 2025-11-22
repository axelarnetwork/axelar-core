package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/app"
	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	keeperMock "github.com/axelarnetwork/axelar-core/x/snapshot/keeper/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
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

func TestKeeper_RegisterProxy(t *testing.T) {
	var (
		ctx              sdk.Context
		snapshotKeeper   keeper.Keeper
		principalAddress sdk.ValAddress
		expectedProxy    sdk.AccAddress
		staker           *mockStaker
		bank             *mock.BankKeeperMock
		validators       []staking.ValidatorI
	)

	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		snapSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "snap")
		validators = genValidators(t, 10, 100)
		staker = newMockStaker(validators...)
		principalAddress = funcs.Must(sdk.ValAddressFromBech32(validators[rand.I64Between(0, 10)].GetOperator()))
		expectedProxy = rand.AccAddr()
		bank = &mock.BankKeeperMock{
			SpendableBalanceFunc: func(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
				if addr.Equals(expectedProxy) {
					return sdk.NewCoin("uaxl", math.NewInt(5000000))
				}
				return sdk.NewCoin("uaxl", math.ZeroInt())
			},
		}

		snapshotKeeper = keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("staking"), snapSubspace, staker, bank, &mock.SlasherMock{})
		snapshotKeeper.SetParams(ctx, types.DefaultParams())
	}
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := snapshotKeeper.ActivateProxy(ctx, principalAddress, expectedProxy)

		assert.NoError(t, err)
		proxy, active := snapshotKeeper.GetProxy(ctx, principalAddress)
		assert.True(t, active)
		assert.Equal(t, expectedProxy, proxy)

	}).Repeat(20))

	t.Run("same addresses", testutils.Func(func(t *testing.T) {
		setup()

		err := snapshotKeeper.ActivateProxy(ctx, expectedProxy.Bytes(), expectedProxy)

		assert.Error(t, err)
	}).Repeat(20))

	t.Run("unknown validator", testutils.Func(func(t *testing.T) {
		setup()

		address := rand.ValAddr()
		proxy := rand.AccAddr()
		err := snapshotKeeper.ActivateProxy(ctx, address, proxy)

		assert.Error(t, err)

	}).Repeat(20))

	t.Run("insufficient funds in proxy", testutils.Func(func(t *testing.T) {
		setup()

		bank.SpendableBalanceFunc = func(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
			if addr.Equals(expectedProxy) {
				return sdk.NewCoin("uaxl", math.NewInt(4999999))
			}
			return sdk.NewCoin("uaxl", math.ZeroInt())
		}

		err := snapshotKeeper.ActivateProxy(ctx, principalAddress, expectedProxy)

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
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		snapSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "snap")
		validators = genValidators(t, 10, 100)
		staker = newMockStaker(validators...)
		principalAddress = funcs.Must(sdk.ValAddressFromBech32(validators[rand.I64Between(0, 10)].GetOperator()))
		expectedProxy = rand.AccAddr()

		bank := &mock.BankKeeperMock{
			SpendableBalanceFunc: func(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
				return sdk.NewCoin("uaxl", math.NewInt(5000000))
			},
		}

		snapshotKeeper = keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("staking"), snapSubspace, staker, bank, &mock.SlasherMock{})
		snapshotKeeper.SetParams(ctx, types.DefaultParams())

		if err := snapshotKeeper.ActivateProxy(ctx, principalAddress, expectedProxy); err != nil {
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
			address = funcs.Must(sdk.ValAddressFromBech32(validators[rand.I64Between(0, 10)].GetOperator()))
			if !bytes.Equal(principalAddress, address) {
				break
			}
		}

		principalAddress = address
		err := snapshotKeeper.DeactivateProxy(ctx, principalAddress)

		assert.Error(t, err)

	}).Repeat(20))
}

func TestKeeper(t *testing.T) {
	var (
		ctx     sdk.Context
		k       keeper.Keeper
		staking *mock.StakingKeeperMock
	)

	givenKeeper := Given("snapshot keeper", func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t)).
			WithBlockHeight(rand.PosI64()).
			WithBlockTime(time.Now())
		subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "snap")

		staking = &mock.StakingKeeperMock{}
		k = keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("snapshot"), subspace, staking, &mock.BankKeeperMock{}, &mock.SlasherMock{})
		k.SetParams(ctx, types.DefaultParams())
	})

	t.Run("CreateSnapshot", func(t *testing.T) {
		var (
			candidates []sdk.ValAddress
			filterFunc func(exported.ValidatorI) bool
			weightFunc func(consensusPower math.Uint) math.Uint
			threshold  utils.Threshold
		)

		validators := slices.Expand(func(i int) stakingtypes.ValidatorI {
			addr := rand.ValAddr()

			return &keeperMock.ValidatorIMock{
				GetOperatorFunc:       func() string { return addr.String() },
				GetConsensusPowerFunc: func(_ math.Int) int64 { return int64(100 - i) },
			}
		}, 100)
		validatorMap := slices.ToMap(validators, func(v stakingtypes.ValidatorI) string { return v.GetOperator() })

		whenAllParamsAreGood := When("all params are good", func() {
			candidates = slices.Map(validators, func(v stakingtypes.ValidatorI) sdk.ValAddress {
				return funcs.Must(sdk.ValAddressFromBech32(v.GetOperator()))
			})
			filterFunc = func(v exported.ValidatorI) bool { return true }
			weightFunc = funcs.Identity[math.Uint]
			threshold = utilstestutils.RandThreshold()

			staking.ValidatorFunc = func(_ context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
				return validatorMap[addr.String()], nil
			}
			staking.IterateBondedValidatorsByPowerFunc = func(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
				for i, v := range validators {
					if fn(int64(i), v) {
						return nil
					}
				}
				return nil
			}
			staking.PowerReductionFunc = func(ctx context.Context) math.Int { return math.OneInt() }
		})

		givenKeeper.
			When2(whenAllParamsAreGood).
			Then("should create a valid snapshot", func(t *testing.T) {
				actual, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.NoError(t, err)
				assert.NoError(t, actual.ValidateBasic())
				assert.Equal(t, ctx.BlockHeight(), actual.Height)
				assert.Equal(t, ctx.BlockTime(), actual.Timestamp)

				expectedBondedWeight := math.ZeroUint()
				for addr, v := range validatorMap {
					weight := math.NewUint(uint64(v.GetConsensusPower(math.OneInt())))
					assert.Equal(t, exported.NewParticipant(funcs.Must(sdk.ValAddressFromBech32(v.GetOperator())), weight), actual.Participants[addr])

					expectedBondedWeight = expectedBondedWeight.Add(weight)
				}
				assert.Equal(t, expectedBondedWeight, actual.BondedWeight)
			}).
			Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("candidates is a subset of validators", func() {
				candidates = []sdk.ValAddress{}
				slices.ForEach(validators, func(v stakingtypes.ValidatorI) {
					if rand.Bools(0.5).Next() {
						candidates = append(candidates, funcs.Must(sdk.ValAddressFromBech32(v.GetOperator())))
					}
				})
				threshold = utils.ZeroThreshold
			}).
			Then("should create a snapshot with only candidates passing the filterFunc", func(t *testing.T) {
				actual, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.NoError(t, err)
				assert.NoError(t, actual.ValidateBasic())
				assert.Len(t, actual.Participants, len(candidates))
			}).
			Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("filterFunc filters some candidate out", func() {
				filterFunc = func(v exported.ValidatorI) bool { return v.GetConsensusPower(math.OneInt()) > 90 }
				threshold = utils.ZeroThreshold
			}).
			Then("should create a snapshot with only candidates passing the filterFunc", func(t *testing.T) {
				actual, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.NoError(t, err)
				assert.NoError(t, actual.ValidateBasic())
				assert.Len(t, actual.Participants, 10)
			}).
			Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("weightFunc should translate consensus power to weight", func() {
				weightFunc = func(math.Uint) math.Uint {
					return math.OneUint()
				}
			}).
			Then("should create a snapshot with participants having the correct weights", func(t *testing.T) {
				actual, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.NoError(t, err)
				assert.NoError(t, actual.ValidateBasic())
				assert.True(t, slices.All(maps.Values(actual.Participants), func(p exported.Participant) bool {
					return p.Weight.Equal(math.OneUint())
				}))
				assert.Equal(t, math.OneUint().MulUint64(uint64(len(actual.Participants))), actual.BondedWeight)
			}).
			Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("threshold cannot be met", func() {
				filterFunc = func(v exported.ValidatorI) bool { return v.GetConsensusPower(math.OneInt()) > 90 }
				threshold = utils.NewThreshold(956, 5050)
			}).
			Then("should return an error", func(t *testing.T) {
				_, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.ErrorContains(t, err, "cannot be met")
			}).
			Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("weight func returns zero weights", func() {
				once := &sync.Once{}
				weightFunc = func(w math.Uint) math.Uint {
					once.Do(func() { w = math.ZeroUint() })
					return w
				}
			}).
			Then("don't include validators with zero weight in snapshot", func(t *testing.T) {
				s, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, utils.NewThreshold(9, 10))

				assert.NoError(t, err)
				participantsWithNonZeroWeights := slices.Map(validators[1:], func(v stakingtypes.ValidatorI) sdk.ValAddress {
					return funcs.Must(sdk.ValAddressFromBech32(v.GetOperator()))
				})
				assert.ElementsMatch(t, participantsWithNonZeroWeights, slices.Map(maps.Values(s.Participants),
					func(p exported.Participant) sdk.ValAddress { return p.Address }))
			}).Run(t)

		givenKeeper.
			When2(whenAllParamsAreGood).
			When("candidate is not a validator", func() {
				candidates = []sdk.ValAddress{rand.ValAddr()}
				threshold = utils.ZeroThreshold
			}).
			Then("no participants are selected", func(t *testing.T) {
				_, err := k.CreateSnapshot(ctx, candidates, filterFunc, weightFunc, threshold)

				assert.ErrorContains(t, err, "snapshot cannot have no participant")
			}).
			Run(t)

	})
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
	totalPower math.Int
}

func newMockStaker(validators ...stakingtypes.ValidatorI) *mockStaker {
	staker := &mockStaker{
		make([]stakingtypes.ValidatorI, 0),
		math.ZeroInt(),
	}

	for _, val := range validators {
		staker.validators = append(staker.validators, val)
		staker.totalPower = staker.totalPower.AddRaw(val.GetConsensusPower(sdk.DefaultPowerReduction))
	}

	return staker
}

func (k mockStaker) BondDenom(_ context.Context) (string, error) {
	return "uaxl", nil
}

func (k mockStaker) GetLastTotalPower(_ context.Context) (math.Int, error) {
	return k.totalPower, nil
}

func (k mockStaker) IterateBondedValidatorsByPower(_ context.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
	for i, val := range k.validators {
		if fn(int64(i), val) {
			return nil
		}
	}
	return nil
}

func (k mockStaker) Validator(_ context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
	for _, validator := range k.validators {
		if validator.GetOperator() == addr.String() {
			return validator, nil
		}
	}

	return nil, stakingtypes.ErrNoValidatorFound
}

func (k mockStaker) PowerReduction(_ context.Context) math.Int {
	return sdk.DefaultPowerReduction
}
