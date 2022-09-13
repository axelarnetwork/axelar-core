package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
)

const bondDenom = "test"

func setup() (sdk.Context, Keeper, *mock.StakingKeeperMock, *mock.BankKeeperMock, *mock.SlasherMock) {
	staking := mock.StakingKeeperMock{}
	bank := mock.BankKeeperMock{}
	slasher := mock.SlasherMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snapshot")
	keeper := NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
		&staking,
		&bank,
		&slasher,
	)

	return ctx, keeper, &staking, &bank, &slasher
}

func TestExportGenesis(t *testing.T) {
	ctx, keeper, staking, bank, _ := setup()
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []types.ProxiedValidator{}))

	staking.BondDenomFunc = func(sdk.Context) string {
		return bondDenom
	}
	bank.GetBalanceFunc = func(sdk.Context, sdk.AccAddress, string) sdk.Coin {
		return sdk.NewCoin(bondDenom, sdk.NewInt(types.DefaultParams().MinProxyBalance))
	}
	staking.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
		return stakingtypes.Validator{}
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
	ctx, keeper, _, _, _ := setup()

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
