package axelarnet_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcTransfer "github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestIBCModule(t *testing.T) {
	var (
		ctx       sdk.Context
		ibcModule axelarnet.AxelarnetIBCModule
		k         keeper.Keeper
		n         *mock.NexusMock
		bankK     *mock.BankKeeperMock

		ack       channeltypes.Acknowledgement
		transfer  types.IBCTransfer
		message   exported.GeneralMessage
		transfers []types.IBCTransfer
	)

	const (
		packetSeq = 1
		channelID = "channel-0"
	)

	givenAnIBCModule := Given("given a module", func() {
		encCfg := appParams.MakeEncodingConfig()
		subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey(types.StoreKey), sdk.NewKVStoreKey("tAxelarnetKey"), types.ModuleName)
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		channelK := &mock.ChannelKeeperMock{
			GetNextSequenceSendFunc: func(sdk.Context, string, string) (uint64, bool) {
				return packetSeq, true
			},
		}

		k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.ModuleName), subspace, channelK, &mock.FeegrantKeeperMock{})
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})

		accountK := &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(string) sdk.AccAddress {
				return rand.AccAddr()
			},
		}

		bankK = &mock.BankKeeperMock{
			SendCoinsFunc: func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
				return nil
			},
			SendCoinsFromAccountToModuleFunc: func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
				return nil
			},
			BurnCoinsFunc: func(sdk.Context, string, sdk.Coins) error { return nil },
		}

		scopeKeeper := capabilitykeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(capabilitytypes.StoreKey), sdk.NewKVStoreKey(capabilitytypes.MemStoreKey))
		scopedTransferK := scopeKeeper.ScopeToModule(ibctransfertypes.ModuleName)
		transferSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey(ibctransfertypes.StoreKey), sdk.NewKVStoreKey("tTrasferKey"), ibctransfertypes.ModuleName)

		transferK := ibctransferkeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("transfer"), transferSubspace, &mock.ChannelKeeperMock{}, &mock.ChannelKeeperMock{}, &mock.PortKeeperMock{}, accountK, bankK, scopedTransferK)
		n = &mock.NexusMock{}
		ibcModule = axelarnet.NewAxelarnetIBCModule(ibcTransfer.NewIBCModule(transferK), ibcK, axelarnet.NewRateLimiter(&k, n), n, bankK)
	})

	fungibleTokenPacket := ibctransfertypes.NewFungibleTokenPacketData(rand.Denom(5, 10), "1", rand.AccAddr().String(), rand.AccAddr().String())

	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), packetSeq, ibctransfertypes.PortID, channelID, ibctransfertypes.PortID, channelID, clienttypes.NewHeight(0, 110), 0)

	whenGetValidAckResult := When("get valid acknowledgement result", func() {
		ack = channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	})

	whenGetValidAckError := When("get valid acknowledgement error", func() {
		ack = channeltypes.NewErrorAcknowledgement(fmt.Errorf("error"))
	})

	whenPendingTransfersExist := When("pending transfers exist", func() {
		transfers = slices.Expand(
			func(_ int) types.IBCTransfer { return testutils.RandomIBCTransfer() },
			int(rand.I64Between(5, 50)),
		)

		slices.ForEach(transfers, func(t types.IBCTransfer) { funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, t)) })
	})

	seqMapsToID := When("packet seq maps to transfer ID", func() {
		transfer = testutils.RandomIBCTransfer()
		transfer.ChannelID = channelID
		funcs.MustNoErr(k.SetSeqIDMapping(ctx, transfer))
		funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, transfer))
	})

	seqMapsToMessageID := When("packet seq maps to message ID", func() {
		message = nexustestutils.RandomMessage(exported.Processing)
		funcs.MustNoErr(k.SetSeqMessageIDMapping(ctx, ibctransfertypes.PortID, channelID, packetSeq, message.ID))

		n.GetMessageFunc = func(sdk.Context, string) (exported.GeneralMessage, bool) { return message, true }
		n.IsAssetRegisteredFunc = func(sdk.Context, exported.Chain, string) bool { return true }
		n.SetMessageFailedFunc = func(ctx sdk.Context, id string) error {
			if id == message.ID {
				message.Status = exported.Failed
			}

			return nil
		}
		n.GetChainByNativeAssetFunc = func(sdk.Context, string) (exported.Chain, bool) { return exported.Chain{}, false }
	})

	whenOnAck := When("on acknowledgement", func() {
		err := ibcModule.OnAcknowledgementPacket(ctx, packet, ack.Acknowledgement(), nil)
		assert.NoError(t, err)
	})

	whenOnTimeout := When("on timeout", func() {
		err := ibcModule.OnTimeoutPacket(ctx, packet, nil)
		assert.NoError(t, err)
	})

	shouldNotChangeTransferState := Then("should not change transfers status", func(t *testing.T) {
		assert.True(t, slices.All(transfers, func(t types.IBCTransfer) bool {
			return funcs.MustOk(k.GetTransfer(ctx, t.ID)).Status == types.TransferPending
		}))
	})

	whenChainIsActivated := When("chain is activated", func() {
		n.GetChainFunc = func(ctx sdk.Context, chain exported.ChainName) (exported.Chain, bool) { return exported.Chain{}, true }
		n.IsChainActivatedFunc = func(ctx sdk.Context, chain exported.Chain) bool { return true }
		n.RateLimitTransferFunc = func(ctx sdk.Context, chain exported.ChainName, asset sdk.Coin, direction exported.TransferDirection) error {
			return nil
		}
	})

	givenAnIBCModule.
		Branch(
			whenGetValidAckResult.
				When2(seqMapsToID).
				When2(whenOnAck).
				Then("should set transfer to complete", func(t *testing.T) {
					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
					assert.Equal(t, types.TransferCompleted, transfer.Status)
				}),

			whenGetValidAckError.
				When2(whenChainIsActivated).
				When2(seqMapsToID).
				When2(whenOnAck).
				Then("should set transfer to failed", func(t *testing.T) {
					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
					assert.Equal(t, types.TransferFailed, transfer.Status)
				}),

			whenPendingTransfersExist.
				When("get invalid ack", func() {
					err := ibcModule.OnAcknowledgementPacket(ctx, packet, rand.BytesBetween(1, 50), nil)
					assert.Error(t, err)
				}).
				Then2(shouldNotChangeTransferState),

			whenGetValidAckResult.
				When2(whenPendingTransfersExist).
				When("seq is not mapped to id", func() {}).
				When2(whenOnAck).
				Then2(shouldNotChangeTransferState),

			seqMapsToID.
				When2(whenChainIsActivated).
				When2(whenOnTimeout).
				Then("should set transfer to failed", func(t *testing.T) {
					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
					assert.Equal(t, types.TransferFailed, transfer.Status)
				}),

			whenPendingTransfersExist.
				When("seq is not mapped to id", func() {}).
				When2(whenChainIsActivated).
				When2(whenOnTimeout).
				Then2(shouldNotChangeTransferState),

			whenGetValidAckError.
				When2(whenChainIsActivated).
				When2(seqMapsToMessageID).
				When2(whenOnAck).
				Then("should set message to failed", func(t *testing.T) {
					assert.Equal(t, exported.Failed, message.Status)
					assert.Len(t, bankK.SendCoinsFromAccountToModuleCalls(), 1)
				}),
		).Run(t)
}
