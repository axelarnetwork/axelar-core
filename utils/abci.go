package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-errors/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

//go:generate moq -out ./mock/abci.go -pkg mock . Logger

// Logger wraps keepers which expose a Logger method
type Logger interface {
	Logger(ctx sdk.Context) log.Logger
}

// RunEndBlocker wraps the given EndBlocker and handles error/panic
func RunEndBlocker(c sdk.Context, l Logger, endBlocker func(sdk.Context) ([]abci.ValidatorUpdate, error)) []abci.ValidatorUpdate {
	ctx, writeCache := c.CacheContext()

	defer func() {
		if r := recover(); r != nil {
			l.Logger(ctx).Error(fmt.Sprintf("panicked running end blocker due to error %v", r))
			l.Logger(ctx).Debug(errors.Wrap(r, 1).ErrorStack())
		}
	}()

	updates, err := endBlocker(ctx)
	if err != nil {
		l.Logger(ctx).Debug(fmt.Sprintf("failed running end blocker due to error %s", err.Error()))
		return nil
	}

	writeCache()
	c.EventManager().EmitEvents(ctx.EventManager().Events())

	return updates
}
