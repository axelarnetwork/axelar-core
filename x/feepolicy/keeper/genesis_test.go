package keeper

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	appparams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
)

func setupKeeper(t *testing.T) (sdk.Context, Keeper) {
	encCfg := appparams.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("feepolicy"), store.NewKVStoreKey("tfeepolicy"), types.ModuleName)

	return ctx, NewKeeper(subspace)
}

func TestInitAndExportGenesis(t *testing.T) {
	ctx, k := setupKeeper(t)

	expected := types.NewGenesisState(types.Params{AllowedFeeDenoms: []string{"uaxl", "uusdc"}})
	k.InitGenesis(ctx, expected)

	actual := k.ExportGenesis(ctx)

	assert.Equal(t, expected, actual)
	assert.NoError(t, actual.Validate())
}

func TestExportGenesis_Default(t *testing.T) {
	ctx, k := setupKeeper(t)
	k.InitGenesis(ctx, types.DefaultGenesisState())

	actual := k.ExportGenesis(ctx)

	assert.Equal(t, types.DefaultGenesisState(), actual)
	assert.NoError(t, actual.Validate())
}
