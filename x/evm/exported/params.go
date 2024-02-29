package exported

import (
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	ModuleName = "evm"
)

var (
	// Ethereum defines properties of the Ethereum chain
	Ethereum = exported.Chain{
		Name:                  "Ethereum",
		SupportsForeignAssets: true,
		KeyType:               tss.Multisig,
		Module:                ModuleName,
	}
)
