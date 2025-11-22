package utils

import (
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-errors/errors"
)

//go:generate moq -out ./mock/abci.go -pkg mock . Logger

// Logger wraps keepers which expose a Logger method
type Logger interface {
	Logger(ctx sdk.Context) log.Logger
}

// RunCached wraps the given function, handles error/panic and rolls back the state if necessary
func RunCached[T any](c sdk.Context, l Logger, f func(sdk.Context) (T, error)) T {
	ctx, writeCache := c.CacheContext()

	defer func() {
		if r := recover(); r != nil {
			l.Logger(ctx).Error(fmt.Sprintf("recovered from panic in cached context: %v", r))
			l.Logger(ctx).Error(string(errors.Wrap(r, 1).Stack()))
		}
	}()

	result, err := f(ctx)
	if err != nil {
		l.Logger(ctx).Debug(fmt.Sprintf("recovered from error in cached context: %s", err.Error()))
		return *new(T)
	}

	writeCache()

	return result
}
