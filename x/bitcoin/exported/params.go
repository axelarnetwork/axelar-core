package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

var (
	// NativeAsset is the native asset on Bitcoin
	NativeAsset = "satoshi"
	// Bitcoin defines properties of the Bitcoin chain
	Bitcoin = exported.Chain{
		Name:                  "Bitcoin",
		SupportsForeignAssets: false,
		KeyType:               tss.Threshold,
		Module:                "bitcoin", // cannot use constant due to import cycle
	}
)
