package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Chain returns a random nexus chain
func Chain() exported.Chain {
	return exported.Chain{
		Name:                  exported.ChainName(rand.NormalizedStrBetween(5, 20)),
		Module:                rand.NormalizedStrBetween(5, 20),
		SupportsForeignAssets: true,
	}
}
