package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
)

func TestMsgServerActivateDeactivateWasm(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	snap := mock.SnapshotterMock{}
	slashing := mock.SlashingKeeperMock{}
	staking := mock.StakingKeeperMock{}
	ax := mock.AxelarnetKeeperMock{}

	msgServer := keeper.NewMsgServerImpl(k, &snap, &slashing, &staking, &ax)

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	assert.True(t, k.IsWasmConnectionActivated(ctx))

	_, err := msgServer.DeactivateChain(sdk.WrapSDKContext(ctx), &types.DeactivateChainRequest{Chains: []nexus.ChainName{":all:"}})
	assert.NoError(t, err)
	assert.False(t, k.IsWasmConnectionActivated(ctx))

	_, err = msgServer.ActivateChain(sdk.WrapSDKContext(ctx), &types.ActivateChainRequest{Chains: []nexus.ChainName{":wasm:"}})
	assert.NoError(t, err)
	assert.True(t, k.IsWasmConnectionActivated(ctx))

	_, err = msgServer.DeactivateChain(sdk.WrapSDKContext(ctx), &types.DeactivateChainRequest{Chains: []nexus.ChainName{"not_wasm"}})
	assert.NoError(t, err)
	assert.True(t, k.IsWasmConnectionActivated(ctx))
}
