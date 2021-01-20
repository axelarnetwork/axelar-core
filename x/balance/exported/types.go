package exported

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DO NOT CHANGE THE ORDER OF THE CHAINS. MODULES WOULD ATTRIBUTE TXS TO THE WRONG CHAIN
	NONE Chain = iota
	Bitcoin
	Ethereum

	// increment when adding a new chain
	ConnectedChainCount = 3
)

var (
	// add labels when new chains are added IN THE CORRECT ORDER
	labels = [ConnectedChainCount]string{"unknown chain", "Bitcoin", "Ethereum"}
)

type Chain int

func (c Chain) Validate() error {
	if c <= 0 || c >= ConnectedChainCount {
		return fmt.Errorf("unknown chain")
	}
	return nil
}

func (c Chain) String() string {
	if c.Validate() == nil {
		return labels[c]
	}
	return labels[0]
}

func ChainFromString(chain string) Chain {
	for i, label := range labels {
		if strings.EqualFold(chain, label) {
			return Chain(i)
		}
	}
	return NONE
}

type CrossChainAddress struct {
	Chain   Chain
	Address string
}

func (a CrossChainAddress) Validate() error {
	if err := a.Chain.Validate(); err != nil {
		return err
	}
	if a.Address == "" {
		return fmt.Errorf("missing address")
	}
	return nil
}

func (a CrossChainAddress) String() string {
	return fmt.Sprintf("chain: %s, address: %s", a.Chain.String(), a.Address)
}

type CrossChainTransfer struct {
	Recipient CrossChainAddress
	Amount    sdk.Coin
	ID        uint64
}
