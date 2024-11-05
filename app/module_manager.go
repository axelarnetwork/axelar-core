package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/axelarnetwork/utils/slices"
)

type FilteredModuleManager struct {
	*module.Manager
	filteredModules []string
}

func NewFilteredModuleManager(appModules []module.AppModule, filteredModules []string) *FilteredModuleManager {
	manager := module.NewManager(appModules...)

	return &FilteredModuleManager{
		manager,
		filteredModules,
	}
}

// RegisterRoutes registers all module routes and module querier routes
func (m *FilteredModuleManager) RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter, legacyQuerierCdc *codec.LegacyAmino) {
	for _, module := range m.Modules {
		if m.isModuleFiltered(module.Name()) {
			continue
		}

		if r := module.Route(); !r.Empty() {
			router.AddRoute(r)
		}
		if r := module.QuerierRoute(); r != "" {
			queryRouter.AddRoute(r, module.LegacyQuerierHandler(legacyQuerierCdc))
		}

	}
}

// RegisterServices registers all module services
func (m *FilteredModuleManager) RegisterServices(cfg module.Configurator) {
	for _, module := range m.Modules {
		if m.isModuleFiltered(module.Name()) {
			continue
		}

		module.RegisterServices(cfg)
	}
}

func (m *FilteredModuleManager) isModuleFiltered(moduleName string) bool {
	return slices.Any(m.filteredModules, func(s string) bool {
		return s == moduleName
	})
}
