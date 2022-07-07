package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	mock2 "github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (codec.Codec, sdk.Context, keeper.Keeper) {
	encCfg := params.MakeEncodingConfig()
	encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&evmTypes.SigMetadata{},
	)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	subspace := paramsTypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")

	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("tss"), subspace, &mock2.SlasherMock{}, &mock.RewarderMock{})
	k.SetParams(ctx, types.DefaultParams())

	return encCfg.Codec, ctx, k
}

func TestGetMigrationHandler(t *testing.T) {
	var (
		cdc         codec.Codec
		ctx         sdk.Context
		k           keeper.Keeper
		handler     func(ctx sdk.Context) error
		sigID       string
		sigMetadata evmTypes.SigMetadata
	)

	givenMigrationHandler := Given("the migration handler", func() {
		cdc, ctx, k = setup()
		handler = keeper.GetMigrationHandler(k)
	})

	givenMigrationHandler.
		When("some sign info with valid EVM sig metadata exists", func() {
			sigMetadata = evmTypes.SigMetadata{
				Type:  evmTypes.SigCommand,
				Chain: nexus.ChainName(rand.NormalizedStr(5)),
			}

			sigID = rand.HexStr(64)
			signInfo := exported.SignInfo{
				SigID:    sigID,
				Metadata: string(evmTypes.ModuleCdc.MustMarshalJSON(&sigMetadata)),
			}
			k.SetInfoForSig(ctx, signInfo.SigID, signInfo)
		}).
		Then("should migrate metadata to module metadata", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			actualSignInfo, ok := k.GetInfoForSig(ctx, sigID)
			assert.True(t, ok)
			assert.Empty(t, actualSignInfo.Metadata)

			var actualSigMetadata evmTypes.SigMetadata
			err = cdc.Unmarshal(actualSignInfo.GetModuleMetadata().Value, &actualSigMetadata)
			assert.NoError(t, err)
			assert.Equal(t, sigMetadata, actualSigMetadata)
		}).
		Run(t)

	givenMigrationHandler.
		When("some sign info with invalid EVM sig metadata exists", func() {
			sigID = rand.HexStr(64)
			signInfo := exported.SignInfo{
				SigID:    sigID,
				Metadata: rand.Str(100),
			}
			k.SetInfoForSig(ctx, signInfo.SigID, signInfo)
		}).
		Then("should ignore", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			actualSignInfo, ok := k.GetInfoForSig(ctx, sigID)
			assert.True(t, ok)
			assert.NotEmpty(t, actualSignInfo.Metadata)
			assert.Nil(t, actualSignInfo.GetModuleMetadata())
		}).
		Run(t)
}
