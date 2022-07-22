package axelarnet

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v2/testing"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (*fakeMock.MultiStoreMock, sdk.Context) {
	store := &fakeMock.MultiStoreMock{}
	cacheStore := &fakeMock.CacheMultiStoreMock{
		WriteFunc: func() {},
	}
	store.CacheMultiStoreFunc = func() sdk.CacheMultiStore { return cacheStore }
	ctx := sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(rand.I64Between(10, 100))

	return store, ctx
}
func TestEndBlocker(t *testing.T) {
	var (
		keeper            *mock.BaseKeeperMock
		transferKeeper    *mock.IBCTransferKeeperMock
		channelKeeper     *mock.ChannelKeeperMock
		transferQueue     *utilsMock.KVQueueMock
		queueSize         int
		queueIdx          int
		ibcTransferErrors int
	)

	store := &fakeMock.MultiStoreMock{}
	cacheStore := &fakeMock.CacheMultiStoreMock{
		WriteFunc: func() {},
	}
	store.CacheMultiStoreFunc = func() sdk.CacheMultiStore { return cacheStore }

	ctx := sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(rand.I64Between(10, 100))
	repeats := 20

	givenTransferQueue := Given("transfer queue", func() {
		store, ctx = setup()

		keeper = &mock.BaseKeeperMock{
			LoggerFunc:                func(ctx sdk.Context) log.Logger { return log.NewNopLogger() },
			GetIBCTransferQueueFunc:   func(ctx sdk.Context) utils.KVQueue { return transferQueue },
			GetRouteTimeoutWindowFunc: func(ctx sdk.Context) uint64 { return 10 },
			EnqueueTransferFunc: func(ctx sdk.Context, transfer types.IBCTransfer) error {
				return nil
			},
		}
		transferKeeper = &mock.IBCTransferKeeperMock{
			SendTransferFunc: func(ctx sdk.Context, sourcePort string, sourceChannel string, token sdk.Coin, sender sdk.AccAddress, receiver string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) error {
				if queueIdx <= ibcTransferErrors {
					return fmt.Errorf("failed to send transfer")
				}

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						ibcchanneltypes.EventTypeSendPacket,
					))
				return nil
			},
		}
		channelKeeper = &mock.ChannelKeeperMock{
			GetChannelClientStateFunc: func(sdk.Context, string, string) (string, ibcclient.ClientState, error) {
				return "07-tendermint-0", clientState(), nil
			},
		}
		transferQueue = &utilsMock.KVQueueMock{
			IsEmptyFunc: func() bool {
				return queueIdx == queueSize
			},
			DequeueFunc: func(value codec.ProtoMarshaler) bool {
				if queueIdx == queueSize {
					return false
				}

				transfer := randomIBCTransfer()
				bz, _ := transfer.Marshal()
				if err := value.Unmarshal(bz); err != nil {
					panic(err)
				}
				queueIdx++

				return true
			},
		}

		queueIdx = 0
		ibcTransferErrors = 0
	})

	givenTransferQueue.
		When("queue is empty", func() {
			queueSize = 0
		}).
		Then("should do nothing", func(t *testing.T) {
			EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight()}, keeper, transferKeeper, channelKeeper)
			assert.Equal(t, len(transferQueue.DequeueCalls()), 0)
			assert.Equal(t, len(transferKeeper.SendTransferCalls()), 0)
		}).
		Run(t, repeats)

	givenTransferQueue.
		When("given a queue size", func() {
			queueSize = int(rand.I64Between(50, 200))
		}).
		Then("should init ibc transfers", func(t *testing.T) {
			EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight()}, keeper, transferKeeper, channelKeeper)
			assert.Equal(t, queueSize, len(transferQueue.DequeueCalls()))
			assert.Equal(t, queueSize, len(transferKeeper.SendTransferCalls()))
			assert.Equal(t, queueSize, slices.Reduce(ctx.EventManager().Events().ToABCIEvents(), 0, func(c int, e abci.Event) int {
				if e.Type == ibcchanneltypes.EventTypeSendPacket {
					c++
				}
				return c
			}))
		}).
		Run(t, repeats)

	givenTransferQueue.
		When("given a queue size and some ibc transfer errors", func() {
			queueSize = int(rand.I64Between(50, 200))
			ibcTransferErrors = int(rand.I64Between(1, int64(queueSize)) + 1)
		}).
		Then("should requeue failed transfers", func(t *testing.T) {
			EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight()}, keeper, transferKeeper, channelKeeper)
			assert.Equal(t, queueSize, len(transferQueue.DequeueCalls()))
			assert.Equal(t, queueSize, len(transferKeeper.SendTransferCalls()))
			assert.Equal(t, queueSize-ibcTransferErrors, slices.Reduce(ctx.EventManager().Events().ToABCIEvents(), 0, func(c int, e abci.Event) int {
				if e.Type == ibcchanneltypes.EventTypeSendPacket {
					c++
				}
				return c
			}))
			fmt.Println(ibcTransferErrors, queueSize)
			assert.Equal(t, ibcTransferErrors, len(keeper.EnqueueTransferCalls()))
		}).
		Run(t, repeats)
}

func randomIBCTransfer() types.IBCTransfer {
	denom := rand.Strings(5, 20).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")).Next()
	return types.IBCTransfer{
		Sender:    rand.AccAddr(),
		Receiver:  rand.NormalizedStrBetween(5, 20),
		Token:     sdk.NewCoin(denom, sdk.NewInt(rand.PosI64())),
		PortID:    rand.NormalizedStrBetween(5, 20),
		ChannelID: rand.NormalizedStrBetween(5, 20),
		ID:        nexus.TransferID(uint64(rand.PosI64())),
	}
}

func clientState() *ibctmtypes.ClientState {
	return ibctmtypes.NewClientState(
		"07-tendermint-0",
		ibctmtypes.DefaultTrustLevel,
		time.Hour*24*7*2,
		time.Hour*24*7*3,
		time.Second*10,
		clienttypes.NewHeight(0, 5),
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
}
