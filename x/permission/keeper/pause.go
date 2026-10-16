package keeper

import (
	"context"
	"strconv"

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

// SkipsPause returns true if the chain is paused and this node was started to skip the pause from this height on
func (k Keeper) SkipsPause(ctx sdk.Context) bool {
	return k.skipPauseHeight != 0 && ctx.BlockHeight() >= k.skipPauseHeight && k.GetPauseLevel(ctx).IsPaused()
}

// IsPauseSkipped returns true if the pause is not enforced. During block execution that is decided by the block's
// pause skip marker, so all nodes agree. Everywhere else (CheckTx, proposals, simulation) by this node's flag.
func (k Keeper) IsPauseSkipped(ctx sdk.Context) bool {
	if ctx.ExecMode() == sdk.ExecModeFinalize {
		return ctx.TransientStore(k.tstoreKey).Has(pauseSkippedKey)
	}

	return k.SkipsPause(ctx)
}

// PreBlock skips the pause for the current block if its first tx is the pause skip marker
func (k Keeper) PreBlock(ctx sdk.Context, txs [][]byte) {
	level := k.GetPauseLevel(ctx)
	if !level.IsPaused() || len(txs) == 0 || !types.IsPauseSkipMarker(txs[0]) {
		return
	}

	ctx.TransientStore(k.tstoreKey).Set(pauseSkippedKey, []byte{1})

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeEmergencyPauseSkipped,
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(ctx.BlockHeight(), 10)),
		sdk.NewAttribute(types.AttributeKeyPauseLevel, level.String()),
		sdk.NewAttribute(types.AttributeKeyProposer, sdk.ConsAddress(ctx.BlockHeader().ProposerAddress).String()),
	))
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	level := c.keeper.GetPauseLevel(sdkCtx)
	if exported.IsAllowedWhilePaused(level, typeURL) || c.keeper.IsPauseSkipped(sdkCtx) {
		return true, nil
	}

	return false, types.ErrChainPaused.Wrapf("message %s is not allowed at pause level %s", typeURL, level)
}
