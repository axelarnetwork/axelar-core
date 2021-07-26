package exported

import "github.com/axelarnetwork/axelar-core/x/nexus/exported"

var (
	// Axelar defines properties of the Axelar chain
	Axelar = exported.Chain{
		Name:                  "Axelar",
		NativeAsset:           "axltest",
		SupportsForeignAssets: true,
	}
)
