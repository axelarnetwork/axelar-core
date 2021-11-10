package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

var (
	// Axelarnet defines properties of the Axelar chain
	Axelarnet = exported.Chain{
		Name:                  "Axelarnet",
		NativeAsset:           "uaxl",
		SupportsForeignAssets: true,
		KeyType:               tss.Multisig,
	}
)
