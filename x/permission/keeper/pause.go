package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/cosmos/gogoproto/types"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

// GetPauseLevel returns the current emergency pause level. PAUSE_LEVEL_UNSPECIFIED means the chain is not paused.
func (k Keeper) GetPauseLevel(ctx sdk.Context) exported.PauseLevel {
	var level gogoprototypes.Int32Value
	if !k.getStore(ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())).Get(pauseLevelKey, &level) {
		return exported.PAUSE_LEVEL_UNSPECIFIED
	}

	return exported.PauseLevel(level.Value)
}

func (k Keeper) setPauseLevel(ctx sdk.Context, level exported.PauseLevel) {
	if !level.IsPaused() {
		k.getStore(ctx).Delete(pauseLevelKey)
		return
	}

	k.getStore(ctx).Set(pauseLevelKey, &gogoprototypes.Int32Value{Value: int32(level)})
}

// CircuitBreaker rejects the execution of messages that are not allowed at the current pause level.
type CircuitBreaker struct {
	keeper Keeper
}

// NewCircuitBreaker is the constructor for CircuitBreaker
func NewCircuitBreaker(k Keeper) CircuitBreaker {
	return CircuitBreaker{keeper: k}
}

// IsAllowed implements baseapp.CircuitBreaker
func (c CircuitBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	level := c.keeper.GetPauseLevel(sdk.UnwrapSDKContext(ctx))
	if exported.IsAllowedWhilePaused(level, typeURL) {
		return true, nil
	}

	return false, types.ErrChainPaused.Wrapf("message %s is not allowed at pause level %s", typeURL, level)
}
