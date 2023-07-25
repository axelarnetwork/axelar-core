package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	// ModuleName exposes axelarnet module name
	ModuleName = "axelarnet"
)

var (
	// NativeAsset is the native asset on Axelarnet
	NativeAsset = "uaxl"

	// Axelarnet defines properties of the Axelar chain
	Axelarnet = exported.Chain{
		Name:                  "Axelarnet",
		SupportsForeignAssets: true,
		KeyType:               tss.None,
		Module:                ModuleName,
	}
)
