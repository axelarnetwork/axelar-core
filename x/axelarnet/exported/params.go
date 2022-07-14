package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (
	// NativeAsset is the native asset on Axelarnet
	NativeAsset = "uaxl"
	// Axelarnet defines properties of the Axelar chain
	Axelarnet = exported.Chain{
		Name:                  "Axelarnet",
		SupportsForeignAssets: true,
		Module:                "axelarnet", // cannot use constant due to import cycle
	}
)
