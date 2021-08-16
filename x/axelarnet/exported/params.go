package exported

import "github.com/axelarnetwork/axelar-core/x/nexus/exported"

var (
	// Axelarnet defines properties of the Axelar chain
	Axelarnet = exported.Chain{
		Name:                  "Axelarnet",
		NativeAsset:           "uaxl",
		SupportsForeignAssets: true,
	}
)
