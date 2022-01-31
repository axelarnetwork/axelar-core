package keeper

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func setup() (sdk.Context, types.BaseKeeper, string) {
	encCfg := params.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)
	chain := "Ethereum"

	return ctx, keeper, chain
}

func TestGetMigrationHandler_updateBytecode(t *testing.T) {
	ctx, keeper, chain := setup()
	nexus := mock.NexusMock{
		GetChainsFunc: func(_ sdk.Context) []exported.Chain {
			return []exported.Chain{{
				Name:   chain,
				Module: types.ModuleName,
			}}
		},
	}

	params := testutils.RandomParams()
	keeper.ForChain(chain).SetParams(ctx, params)

	handler := GetMigrationHandler(keeper, &nexus)
	handler(ctx)

	actual := keeper.ForChain(chain).GetParams(ctx)
	assert.Equal(t, actual.GatewayCode, types.DefaultParams()[0].GatewayCode)
	assert.Equal(t, actual.TokenCode, types.DefaultParams()[0].TokenCode)
	assert.Equal(t, actual.Burnable, types.DefaultParams()[0].Burnable)
}
