package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func setup() (sdk.Context, types.BaseKeeper) {
	encCfg := params.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)

	return ctx, keeper
}

func TestGetMigrationHandler_deletePendingGateway(t *testing.T) {
	ctx, keeper := setup()
	evmChains := []exported.Chain{
		{
			Name:   "evm-1",
			Module: types.ModuleName,
		},
		{
			Name:   "evm-2",
			Module: types.ModuleName,
		},
		{
			Name:   "evm-3",
			Module: types.ModuleName,
		},
	}
	nexus := mock.NexusMock{
		GetChainsFunc: func(_ sdk.Context) []exported.Chain {
			return evmChains
		},
	}
	keeper.ForChain(evmChains[0].Name).(chainKeeper).setGateway(ctx, types.Gateway{Address: types.Address(common.BytesToAddress(rand.Bytes(20)))})
	keeper.ForChain(evmChains[1].Name).(chainKeeper).setGateway(ctx, types.Gateway{Address: types.Address(common.BytesToAddress(rand.Bytes(20))), Status: types.GatewayStatusConfirmed})
	keeper.ForChain(evmChains[2].Name).(chainKeeper).setGateway(ctx, types.Gateway{Address: types.Address(common.BytesToAddress(rand.Bytes(20))), Status: types.GatewayStatusPending})

	for _, chain := range evmChains {
		keeper.ForChain(chain.Name).SetParams(ctx, testutils.RandomParams())
	}

	handler := GetMigrationHandler(keeper, &nexus)
	handler(ctx)

	_, actual := keeper.ForChain(evmChains[0].Name).GetGatewayAddress(ctx)
	assert.True(t, actual)
	_, actual = keeper.ForChain(evmChains[1].Name).GetGatewayAddress(ctx)
	assert.True(t, actual)
	_, actual = keeper.ForChain(evmChains[2].Name).GetGatewayAddress(ctx)
	assert.False(t, actual)
}
