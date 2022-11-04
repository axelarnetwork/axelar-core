package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// RandomChain returns a random nexus chain
func RandomChain() exported.Chain {
	return exported.Chain{
		Name:                  RandomChainName(),
		Module:                rand.NormalizedStrBetween(5, 20),
		SupportsForeignAssets: true,
	}
}

// RandomChainName generates a random chain name
func RandomChainName() exported.ChainName {
	return exported.ChainName(rand.NormalizedStrBetween(5, 20))
}

// RandomTransferID generates a random transfer ID
func RandomTransferID() exported.TransferID {
	return exported.TransferID(rand.PosI64())
}

// RandomDirection generates a random transfer direction
func RandomDirection() exported.TransferDirection {
	return exported.TransferDirection(rand.I64Between(1, int64(len(exported.TransferDirection_name))))
}
