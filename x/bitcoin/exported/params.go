package exported

import "github.com/axelarnetwork/axelar-core/x/balance/exported"

var (
	// Bitcoin defines properties of the Bitcoin chain
	Bitcoin = exported.Chain{
		Name:                  "Bitcoin",
		NativeAsset:           "satoshi",
		SupportsForeignAssets: false,
	}
)
