package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// RandomChain returns a random nexus chain
func RandomChain(keyTypes ...tss.KeyType) exported.Chain {
	if len(keyTypes) == 0 {
		keyTypes = []tss.KeyType{tss.None, tss.Multisig}
	}
	return exported.Chain{
		Name:                  RandomChainName(),
		Module:                rand.NormalizedStrBetween(5, 20),
		SupportsForeignAssets: true,
		KeyType:               rand.Of(keyTypes...),
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
