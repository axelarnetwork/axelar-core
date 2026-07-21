package app

import (
	"errors"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators is intentionally unsupported.
//
// axelar-core performs upgrades exclusively via in-place store migrations (x/upgrade),
// never by exporting and re-importing genesis.
func (app *AxelarApp) ExportAppStateAndValidators(_ bool, _ []string, _ []string) (servertypes.ExportedApp, error) {
	return servertypes.ExportedApp{}, errors.New("state export is not supported: axelar-core upgrades via in-place store migrations, not genesis export/import")
}
