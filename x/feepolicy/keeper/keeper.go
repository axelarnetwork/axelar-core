package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
)

// Keeper provides access to the feepolicy module's parameters.
type Keeper struct {
	paramSpace paramtypes.Subspace
}

// NewKeeper returns a new feepolicy keeper.
func NewKeeper(paramSpace paramtypes.Subspace) Keeper {
	return Keeper{paramSpace: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

// SetParams sets the module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetAllowedFeeDenoms returns the denominations that may be used to pay tx fees.
func (k Keeper) GetAllowedFeeDenoms(ctx sdk.Context) []string {
	return k.GetParams(ctx).AllowedFeeDenoms
}
