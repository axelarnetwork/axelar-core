package ante_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	abciproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/ante"
)

func TestLimitSimulationGasDecorator_AnteHandle_WithBlockGasLimit(t *testing.T) {
	anteHandler := sdk.ChainAnteDecorators(ante.NewLimitSimulationGasDecorator(nil))

	ctx := sdk.NewContext(fake.NewMultiStore(), abciproto.Header{}, true, log.NewTestLogger(t)).
		WithConsensusParams(abciproto.ConsensusParams{Block: &abciproto.BlockParams{MaxGas: 1000}}).
		WithGasMeter(store.NewInfiniteGasMeter())

	ctx.GasMeter().ConsumeGas(100, "test")

	ctx, err := anteHandler(ctx, nil, false)
	assert.NoError(t, err)
	assert.EqualValues(t, 100, ctx.GasMeter().GasConsumed())

	// handler should not replace the gas meter, so there should be no limit
	assert.NotPanics(t, func() {
		ctx.GasMeter().ConsumeGas(2000, "test")
	})

	ctx = ctx.WithGasMeter(store.NewInfiniteGasMeter())
	ctx.GasMeter().ConsumeGas(100, "test")

	ctx, err = anteHandler(ctx, nil, true)
	assert.NoError(t, err)
	assert.EqualValues(t, 100, ctx.GasMeter().GasConsumed())

	// handler should have replaced the gas meter, now 1000 should be the limit
	assert.False(t, ctx.GasMeter().IsOutOfGas())
	assert.Panics(t, func() {
		ctx.GasMeter().ConsumeGas(2000, "test")
	})
	assert.True(t, ctx.GasMeter().IsOutOfGas())
}

func TestLimitSimulationGasDecorator_AnteHandle_WithoutBlockGasLimit(t *testing.T) {
	anteHandler := sdk.ChainAnteDecorators(ante.NewLimitSimulationGasDecorator(nil))

	ctx := sdk.NewContext(fake.NewMultiStore(), abciproto.Header{}, true, log.NewTestLogger(t)).
		WithConsensusParams(abciproto.ConsensusParams{Block: &abciproto.BlockParams{MaxGas: 0}}).
		WithGasMeter(store.NewInfiniteGasMeter())

	ctx.GasMeter().ConsumeGas(100, "test")

	ctx, err := anteHandler(ctx, nil, false)
	assert.NoError(t, err)
	assert.EqualValues(t, 100, ctx.GasMeter().GasConsumed())

	// handler should not replace the gas meter, so there should be no limit
	assert.NotPanics(t, func() {
		ctx.GasMeter().ConsumeGas(2000, "test")
	})

	ctx = ctx.WithGasMeter(store.NewInfiniteGasMeter())
	ctx.GasMeter().ConsumeGas(100, "test")

	ctx, err = anteHandler(ctx, nil, true)
	assert.NoError(t, err)
	assert.EqualValues(t, 100, ctx.GasMeter().GasConsumed())

	// handler should have replaced the gas meter, but it's still limitless
	assert.False(t, ctx.GasMeter().IsOutOfGas())
	assert.NotPanics(t, func() {
		ctx.GasMeter().ConsumeGas(2000, "test")
	})
	assert.False(t, ctx.GasMeter().IsOutOfGas())
}
