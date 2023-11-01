package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// RandomChain returns a random nexus chain
func RandomChain() exported.Chain {
	return exported.Chain{
		Name:                  RandomChainName(),
		Module:                rand.NormalizedStrBetween(5, 20),
		SupportsForeignAssets: true,
		KeyType:               tss.None,
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

// RandomCrossChainAddress generates a random cross chain address
func RandomCrossChainAddress() exported.CrossChainAddress {
	return exported.CrossChainAddress{
		Chain:   RandomChain(),
		Address: rand.AccAddr().String(),
	}
}

// RandomMessage generates a random message
func RandomMessage(statuses ...exported.GeneralMessage_Status) exported.GeneralMessage {
	if len(statuses) == 0 {
		statuses = []exported.GeneralMessage_Status{exported.Approved, exported.Processing, exported.Executed, exported.Failed}
	}
	coin := rand.Coin()

	msg := exported.NewGeneralMessage(
		rand.StrBetween(10, 20),
		RandomCrossChainAddress(),
		RandomCrossChainAddress(),
		rand.Bytes(32),
		rand.Bytes(32),
		uint64(rand.I64Between(0, 10000)),
		&coin,
	)
	msg.Status = rand.Of(statuses...)

	return msg
}
