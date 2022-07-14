package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (
	// Ethereum defines properties of the Ethereum chain
	Ethereum = exported.Chain{
		Name:                  "Ethereum",
		SupportsForeignAssets: true,
		Module:                "evm", // cannot use constant due to import cycle
	}
)
