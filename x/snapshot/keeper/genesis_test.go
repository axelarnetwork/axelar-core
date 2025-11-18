package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
)

const bondDenom = "test"

func setup(t log.TestingT) (sdk.Context, Keeper, *mock.StakingKeeperMock, *mock.BankKeeperMock, *mock.SlasherMock) {
	staking := mock.StakingKeeperMock{}
	bank := mock.BankKeeperMock{}
	slasher := mock.SlasherMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "snapshot")
	keeper := NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&staking,
		&bank,
		&slasher,
	)

	return ctx, keeper, &staking, &bank, &slasher
}

func TestExportGenesis(t *testing.T) {
	ctx, keeper, staking, bank, _ := setup(t)
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []types.ProxiedValidator{}))

	staking.BondDenomFunc = func(context.Context) (string, error) {
		return bondDenom, nil
	}
	bank.SpendableBalanceFunc = func(context.Context, sdk.AccAddress, string) sdk.Coin {
		return sdk.NewCoin(bondDenom, math.NewInt(types.DefaultParams().MinProxyBalance))
	}
	staking.ValidatorFunc = func(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
		return stakingtypes.Validator{}, nil
	}

	proxiedValidatorCount := rand.I64Between(10, 100)
	validators := make([]sdk.ValAddress, proxiedValidatorCount)
	proxies := make([]sdk.AccAddress, proxiedValidatorCount)
	expectedProxiedValidators := make([]types.ProxiedValidator, proxiedValidatorCount)

	for i := 0; i < int(proxiedValidatorCount); i++ {
		validators[i] = rand.ValAddr()
		proxies[i] = rand.AccAddr()

		active := rand.Bools(0.5).Next()
		expectedProxiedValidators[i] = types.NewProxiedValidator(validators[i], proxies[i], active)

		err := keeper.ActivateProxy(ctx, validators[i], proxies[i])
		assert.NoError(t, err)

		if !active {
			err := keeper.DeactivateProxy(ctx, validators[i])
			assert.NoError(t, err)
		}
	}

	actual := keeper.ExportGenesis(ctx)
	expected := types.NewGenesisState(types.DefaultParams(), expectedProxiedValidators)

	assert.NoError(t, actual.Validate())
	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.ProxiedValidators, actual.ProxiedValidators)
}

func TestInitGenesis(t *testing.T) {
	ctx, keeper, _, _, _ := setup(t)

	proxiedValidatorCount := rand.I64Between(10, 100)
	expectedProxiedValidators := make([]types.ProxiedValidator, proxiedValidatorCount)
	for i := 0; i < int(proxiedValidatorCount); i++ {
		active := rand.Bools(0.5).Next()
		expectedProxiedValidators[i] = types.NewProxiedValidator(rand.ValAddr(), rand.AccAddr(), active)
	}

	expected := types.NewGenesisState(types.DefaultParams(), expectedProxiedValidators)
	keeper.InitGenesis(ctx, expected)
	actual := keeper.ExportGenesis(ctx)

	assert.NoError(t, actual.Validate())
	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.ProxiedValidators, actual.ProxiedValidators)
}
