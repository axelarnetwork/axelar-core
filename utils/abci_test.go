package utils_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/mock"
	testutils "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, utils.Logger, *fakeMock.MultiStoreMock, *fakeMock.CacheMultiStoreMock) {
	store := &fakeMock.MultiStoreMock{}
	cacheStore := &fakeMock.CacheMultiStoreMock{
		WriteFunc: func() {},
	}
	store.CacheMultiStoreFunc = func() sdk.CacheMultiStore { return cacheStore }

	ctx := sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())
	l := &mock.LoggerMock{LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() }}

	return ctx, l, store, cacheStore
}

func TestRunEndBlocker(t *testing.T) {
	t.Run("should recover and return nil if end blocker panics", func(t *testing.T) {
		ctx, l, store, cacheStore := setup()

		actual := utils.RunCached(ctx, l, func(sdk.Context) ([]types.ValidatorUpdate, error) {
			panic(fmt.Errorf("panic"))
		})

		assert.Nil(t, actual)
		assert.Len(t, store.CacheMultiStoreCalls(), 1)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	})

	t.Run("should return nil and not write if end blocker fails", func(t *testing.T) {
		ctx, l, store, cacheStore := setup()

		actual := utils.RunCached(ctx, l, func(sdk.Context) ([]types.ValidatorUpdate, error) {
			return []types.ValidatorUpdate{{}}, fmt.Errorf("error")
		})

		assert.Nil(t, actual)
		assert.Len(t, store.CacheMultiStoreCalls(), 1)
		assert.Len(t, cacheStore.WriteCalls(), 0)
	})

	t.Run("should return updates and write if end blocker succeeds", func(t *testing.T) {
		ctx, l, store, cacheStore := setup()

		expected := []types.ValidatorUpdate{{}}
		actual := utils.RunCached(ctx, l, func(sdk.Context) ([]types.ValidatorUpdate, error) {
			return expected, nil
		})

		assert.Equal(t, expected, actual)
		assert.Len(t, store.CacheMultiStoreCalls(), 1)
		assert.Len(t, cacheStore.WriteCalls(), 1)
	})

	var (
		baseCtx sdk.Context
		logger  utils.Logger
	)
	testutils.Given("a base context with event manager", func() {
		ctx, l, store, _ := setup()
		logger = l
		baseCtx = ctx.
			WithMultiStore(store).
			WithEventManager(sdk.NewEventManager())
	}).When("running an end blocker that emits events", func() {
		utils.RunCached(baseCtx, logger, func(cachedCtx sdk.Context) ([]types.ValidatorUpdate, error) {
			cachedCtx.EventManager().EmitEvent(sdk.Event{})
			return nil, nil
		})
	}).Then("pass events down to the base context", func(t *testing.T) {
		assert.Len(t, baseCtx.EventManager().Events(), 1)
	}).Run(t)
}
