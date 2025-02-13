package distribution

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"

	"github.com/axelarnetwork/axelar-core/x/distribution/keeper"
)

var _ module.AppModule = AppModule{}

type AppModule struct {
	distr.AppModule

	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(distrAppModule distr.AppModule, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModule: distrAppModule,
		keeper:    keeper,
	}
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}
