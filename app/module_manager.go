package app

import (
	"cosmossdk.io/core/appmodule"
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

// RegisterServices registers all module services
func (f *FilteredModuleManager) RegisterServices(cfg module.Configurator) {
	for _, m := range f.Modules {
		if m, ok := m.(module.HasName); ok && f.isModuleFiltered(m.Name()) {
			continue
		}

		// Handle old SDK module interface
		if m, ok := m.(module.HasServices); ok {
			m.RegisterServices(cfg)
		}

		// Handle new Core API module interface (e.g., consensus module)
		if m, ok := m.(appmodule.HasServices); ok {
			if err := m.RegisterServices(cfg); err != nil {
				panic(err)
			}
		}
	}
}

func (f *FilteredModuleManager) isModuleFiltered(moduleName string) bool {
	return slices.Any(f.filteredModules, func(s string) bool {
		return s == moduleName
	})
}
