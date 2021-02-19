package exported

import "github.com/axelarnetwork/axelar-core/x/nexus/exported"

var (
	// Ethereum defines properties of the Ethereum chain
	Ethereum = exported.Chain{
		Name:                  "Ethereum",
		NativeAsset:           "wei",
		SupportsForeignAssets: true,
	}
)
