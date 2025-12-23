package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmkeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
)

func TestMsgServerActivateDeactivateWasm(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	snap := mock.SnapshotterMock{}
	slashing := mock.SlashingKeeperMock{}
	staking := mock.StakingKeeperMock{}
	ax := mock.AxelarnetKeeperMock{}

	msgServer := keeper.NewMsgServerImpl(k, &snap, &slashing, &staking, &ax)

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

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

func TestUpdateParams(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	msgServer := keeper.NewMsgServerImpl(k, &mock.SnapshotterMock{}, &mock.SlashingKeeperMock{}, &mock.StakingKeeperMock{}, &mock.AxelarnetKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

	p := types.DefaultParams()
	p.ChainMaintainerCheckWindow = p.ChainMaintainerCheckWindow + 1
	_, err := msgServer.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: "", Params: p})
	assert.NoError(t, err)
	assert.Equal(t, p, k.GetParams(ctx))
}

func TestRetryFailedMessage(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	validators := types.NewAddressValidators()
	validators.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator())
	validators.Seal()
	k.SetAddressValidators(validators)

	msgServer := keeper.NewMsgServerImpl(k, &mock.SnapshotterMock{}, &mock.SlashingKeeperMock{}, &mock.StakingKeeperMock{}, &mock.AxelarnetKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

	k.SetMessageRouter(types.NewMessageRouter().AddRoute(evm.Ethereum.Module, func(_ sdk.Context, _ nexus.RoutingContext, _ nexus.GeneralMessage) error {
		return nil
	}))

	t.Run("message not found", func(t *testing.T) {
		_, err := msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     "non-existent-id",
		})
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("message not in failed status", func(t *testing.T) {
		msg := nexus.GeneralMessage{
			ID: rand.NormalizedStr(10),
			Sender: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			Recipient: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			PayloadHash:   evmtestutils.RandomHash().Bytes(),
			Status:        nexus.Approved,
			SourceTxID:    evmtestutils.RandomHash().Bytes(),
			SourceTxIndex: 0,
		}

		err := k.SetNewMessage(ctx, msg)
		assert.NoError(t, err)

		_, err = msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     msg.ID,
		})
		assert.ErrorContains(t, err, "not in failed status")
	})

	t.Run("successfully retry failed message", func(t *testing.T) {
		k.SetChain(ctx, evm.Ethereum)
		k.ActivateChain(ctx, evm.Ethereum)

		msg := nexus.GeneralMessage{
			ID: rand.NormalizedStr(10),
			Sender: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			Recipient: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			PayloadHash:   evmtestutils.RandomHash().Bytes(),
			Status:        nexus.Approved,
			SourceTxID:    evmtestutils.RandomHash().Bytes(),
			SourceTxIndex: 0,
		}

		err := k.SetNewMessage(ctx, msg)
		assert.NoError(t, err)

		err = k.RouteMessage(ctx, msg.ID)
		assert.NoError(t, err)

		err = k.SetMessageFailed(ctx, msg.ID)
		assert.NoError(t, err)

		storedMsg, ok := k.GetMessage(ctx, msg.ID)
		assert.True(t, ok)
		assert.Equal(t, nexus.Failed, storedMsg.Status)

		_, err = msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     msg.ID,
		})
		assert.NoError(t, err)

		dequeuedMsg, ok := k.DequeueRouteMessage(ctx)
		assert.True(t, ok)
		assert.Equal(t, msg.ID, dequeuedMsg.ID)
	})
}
