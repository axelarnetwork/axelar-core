package app

import (
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

		if m, ok := m.(module.HasServices); ok {
			m.RegisterServices(cfg)
		}
	}
}

func (f *FilteredModuleManager) isModuleFiltered(moduleName string) bool {
	return slices.Any(f.filteredModules, func(s string) bool {
		return s == moduleName
	})
}
