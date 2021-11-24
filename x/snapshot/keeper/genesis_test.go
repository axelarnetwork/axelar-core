package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	exportedmock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

const bondDenom = "test"

func setup() (sdk.Context, Keeper, *mock.StakingKeeperMock, *mock.BankKeeperMock, *exportedmock.SlasherMock, *exportedmock.TssMock) {
	staking := mock.StakingKeeperMock{}
	bank := mock.BankKeeperMock{}
	slasher := exportedmock.SlasherMock{}
	tss := exportedmock.TssMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Marshaler, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "snapshot")
	keeper := NewKeeper(
		encodingConfig.Marshaler,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
		&staking,
		&bank,
		&slasher,
		&tss,
	)

	return ctx, keeper, &staking, &bank, &slasher, &tss
}

func getRandomSnapshot(counter int64) exported.Snapshot {
	validatorCount := rand.I64Between(1, 100)
	validators := make([]exported.Validator, validatorCount)
	totalShareCount := sdk.ZeroInt()

	for i := 0; i < int(validatorCount); i++ {
		shareCount := rand.I64Between(1, 20)
		validator := stakingtypes.Validator{OperatorAddress: rand.ValAddr().String()}
		validators[i] = exported.NewValidator(&validator, shareCount)
		totalShareCount = totalShareCount.AddRaw(shareCount)
	}

	return exported.NewSnapshot(
		validators,
		time.Time{},
		rand.PosI64(),
		totalShareCount,
		counter,
		tss.WeightedByStake,
		tss.ComputeAbsCorruptionThreshold(tsstypes.DefaultParams().KeyRequirements[0].SafetyThreshold, totalShareCount),
	)
}

func TestExportGenesis(t *testing.T) {
	ctx, keeper, staking, bank, _, _ := setup()
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []exported.Snapshot{}, []types.ValidatorProxy{}))

	staking.BondDenomFunc = func(sdk.Context) string {
		return bondDenom
	}
	bank.GetBalanceFunc = func(sdk.Context, sdk.AccAddress, string) sdk.Coin {
		return sdk.NewCoin(bondDenom, sdk.NewInt(types.DefaultParams().MinProxyBalance))
	}
	staking.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
		return stakingtypes.Validator{}
	}

	validatorProxyCount := rand.I64Between(10, 100)
	validators := make([]sdk.ValAddress, validatorProxyCount)
	proxies := make([]sdk.AccAddress, validatorProxyCount)
	expectedValidatorProxies := make([]types.ValidatorProxy, validatorProxyCount)

	for i := 0; i < int(validatorProxyCount); i++ {
		validators[i] = rand.ValAddr()
		proxies[i] = rand.AccAddr()

		active := rand.Bools(0.5).Next()
		expectedValidatorProxies[i] = types.NewValidatorProxy(validators[i], proxies[i], active)

		err := keeper.RegisterProxy(ctx, validators[i], proxies[i])
		assert.NoError(t, err)

		if !active {
			err := keeper.DeactivateProxy(ctx, validators[i])
			assert.NoError(t, err)
		}
	}

	snapshotCount := rand.I64Between(10, 100)
	expectedSnapshots := make([]exported.Snapshot, snapshotCount)

	for i := 0; i < int(snapshotCount); i++ {
		expectedSnapshots[i] = getRandomSnapshot(int64(i))
		keeper.setSnapshot(ctx, expectedSnapshots[i])
	}
	keeper.setSnapshotCount(ctx, snapshotCount)

	actual := keeper.ExportGenesis(ctx)
	expected := types.NewGenesisState(types.DefaultParams(), expectedSnapshots, expectedValidatorProxies)

	assert.NoError(t, actual.Validate())
	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.Snapshots, actual.Snapshots)
	assert.ElementsMatch(t, expected.ValidatorProxies, actual.ValidatorProxies)
}

func TestInitGenesis(t *testing.T) {
	ctx, keeper, _, _, _, _ := setup()

	snapshotCount := rand.I64Between(10, 100)
	expectedSnapshots := make([]exported.Snapshot, snapshotCount)
	for i := 0; i < int(snapshotCount); i++ {
		expectedSnapshots[i] = getRandomSnapshot(int64(i))
	}

	validatorProxyCount := rand.I64Between(10, 100)
	expectedValidatorProxies := make([]types.ValidatorProxy, validatorProxyCount)
	for i := 0; i < int(validatorProxyCount); i++ {
		active := rand.Bools(0.5).Next()
		expectedValidatorProxies[i] = types.NewValidatorProxy(rand.ValAddr(), rand.AccAddr(), active)
	}

	expected := types.NewGenesisState(types.DefaultParams(), expectedSnapshots, expectedValidatorProxies)
	keeper.InitGenesis(ctx, expected)
	actual := keeper.ExportGenesis(ctx)

	assert.NoError(t, actual.Validate())
	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.Snapshots, actual.Snapshots)
	assert.ElementsMatch(t, expected.ValidatorProxies, actual.ValidatorProxies)
}
