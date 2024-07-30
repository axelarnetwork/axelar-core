package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LimitSimulationGasDecorator ante decorator to limit gas in simulation calls
// This is a fix for cosmwasm's LimitSimulationGasDecorator.
// The original implementation discards any gas consumption before this decorator is called.
type LimitSimulationGasDecorator struct {
	gasLimit *sdk.Gas
}

// NewLimitSimulationGasDecorator constructor accepts nil value to fallback to block gas limit.
func NewLimitSimulationGasDecorator(gasLimit *sdk.Gas) *LimitSimulationGasDecorator {
	if gasLimit != nil && *gasLimit == 0 {
		panic("gas limit must not be zero")
	}
	return &LimitSimulationGasDecorator{gasLimit: gasLimit}
}

// AnteHandle that limits the maximum gas available in simulations only.
// Fixed from the original implementation by carrying over the consumed gas in the discarded gas meter.
func (d LimitSimulationGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !simulate {
		// Wasm code is not executed in checkTX so that we don't need to limit it further.
		// Tendermint rejects the TX afterwards when the tx.gas > max block gas.
		// On deliverTX we rely on the tendermint/sdk mechanics that ensure
		// tx has gas set and gas < max block gas
		return next(ctx, tx, simulate)
	}

	var gasMeter sdk.GasMeter

	limit, hasLimit := d.getGasLimit(ctx)
	if hasLimit {
		gasMeter = sdk.NewGasMeter(limit)
		gasMeter.ConsumeGas(ctx.GasMeter().GasConsumed(), "ante handler")
		ctx = ctx.WithGasMeter(gasMeter)
	}

	return next(ctx, tx, simulate)
}

func (d LimitSimulationGasDecorator) getGasLimit(ctx sdk.Context) (sdk.Gas, bool) {
	// apply custom node gas limit
	if d.gasLimit != nil {
		return *d.gasLimit, true
	}

	// default to max block gas when set, to be on the safe side
	if maxGas := ctx.ConsensusParams().GetBlock().MaxGas; maxGas > 0 {
		return sdk.Gas(maxGas), true
	}

	return 0, false
}
