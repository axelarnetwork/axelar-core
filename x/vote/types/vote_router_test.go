package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/utils/funcs"
	testutils "github.com/axelarnetwork/utils/test"
)

func TestVoteRouter(t *testing.T) {
	var (
		router   VoteRouter
		handler  *mock.VoteHandlerMock
		ctx      sdk.Context
		storeKey = sdk.StoreKey(sdk.NewKVStoreKey("test"))
	)

	withRegisteredHandler := testutils.Given("a vote router", func() {
		router = NewRouter()
	}).
		Given("a context", func() {
			ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		}).
		When("a handler is registered", func() {
			handler = &mock.VoteHandlerMock{}
			router = router.AddHandler("test", handler)
		})

	withRegisteredHandler.Then("the router knows the handler", func(t *testing.T) {
		assert.True(t, router.HasHandler("test"))
	}).Run(t)

	withRegisteredHandler.When("handler changes state and emits event", func() {
		handler.HandleResultFunc = func(ctx sdk.Context, _ codec.ProtoMarshaler) error {
			ctx.KVStore(storeKey).Set([]byte("key1"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return nil
		}

		handler.HandleFailedPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key2"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return nil
		}

		handler.HandleExpiredPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key3"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return nil
		}

		handler.HandleCompletedPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key4"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return nil
		}
	}).
		Then("state is persisted and events emitted", func(t *testing.T) {
			assert.NoError(t, router.GetHandler("test").HandleResult(ctx, nil))
			assert.NoError(t, router.GetHandler("test").HandleExpiredPoll(ctx, nil))
			assert.NoError(t, router.GetHandler("test").HandleCompletedPoll(ctx, nil))
			assert.NoError(t, router.GetHandler("test").HandleFailedPoll(ctx, nil))

			assert.Equal(t, []byte("value"), ctx.KVStore(storeKey).Get([]byte("key1")))
			assert.Equal(t, []byte("value"), ctx.KVStore(storeKey).Get([]byte("key2")))
			assert.Equal(t, []byte("value"), ctx.KVStore(storeKey).Get([]byte("key3")))
			assert.Equal(t, []byte("value"), ctx.KVStore(storeKey).Get([]byte("key4")))
			assert.Len(t, ctx.EventManager().Events(), 4)
		}).Run(t)

	withRegisteredHandler.When("handler returns error", func() {
		handler.HandleResultFunc = func(ctx sdk.Context, _ codec.ProtoMarshaler) error {
			ctx.KVStore(storeKey).Set([]byte("key1"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return fmt.Errorf("some error")
		}

		handler.HandleFailedPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key2"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return fmt.Errorf("some error")
		}

		handler.HandleExpiredPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key3"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return fmt.Errorf("some error")
		}

		handler.HandleCompletedPollFunc = func(ctx sdk.Context, _ exported.Poll) error {
			ctx.KVStore(storeKey).Set([]byte("key4"), []byte("value"))

			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&Voted{}))
			return fmt.Errorf("some error")
		}
	}).
		Then("ctx is rolled back and no events emitted", func(t *testing.T) {
			assert.Error(t, router.GetHandler("test").HandleResult(ctx, nil))
			assert.Error(t, router.GetHandler("test").HandleExpiredPoll(ctx, nil))
			assert.Error(t, router.GetHandler("test").HandleCompletedPoll(ctx, nil))
			assert.Error(t, router.GetHandler("test").HandleFailedPoll(ctx, nil))

			assert.Nil(t, ctx.KVStore(storeKey).Get([]byte("key1")))
			assert.Nil(t, ctx.KVStore(storeKey).Get([]byte("key2")))
			assert.Nil(t, ctx.KVStore(storeKey).Get([]byte("key3")))
			assert.Nil(t, ctx.KVStore(storeKey).Get([]byte("key4")))
			assert.Len(t, ctx.EventManager().Events(), 0)
		}).Run(t)
}
