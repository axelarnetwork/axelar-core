package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, BaseKeeper) {
	encCfg := params.MakeEncodingConfig()

	encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&multisigtypes.MultiSig{},
	)

	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)

	for _, params := range types.DefaultParams() {
		keeper.ForChain(params.Chain).SetParams(ctx, params)
	}

	return ctx, keeper
}

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		keeper  BaseKeeper
		handler func(ctx sdk.Context) error
	)

	evmChains := []nexus.Chain{exported.Ethereum}
	givenMigrationHandler := Given("the migration handler", func() {
		ctx, keeper = setup()
		nexus := mock.NexusMock{
			GetChainsFunc: func(_ sdk.Context) []nexus.Chain {
				return evmChains
			},
		}

		handler = Migrate5To6(keeper, &nexus)
	})

	givenMigrationHandler.
		When("TransferLimit param is not set", func() {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				subspace, _ := ck.getSubspace(ctx)
				subspace.Set(ctx, types.KeyTransferLimit, int64(0))
			}
		}).
		Then("should set TransferLimit param", func(t *testing.T) {
			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.Zero(t, ck.GetParams(ctx).TransferLimit)
			}

			err := handler(ctx)
			assert.NoError(t, err)

			for _, chain := range evmChains {
				ck := keeper.ForChain(chain.Name).(chainKeeper)
				assert.Equal(t, types.DefaultParams()[0].TransferLimit, ck.GetParams(ctx).TransferLimit)
			}
		})
}
