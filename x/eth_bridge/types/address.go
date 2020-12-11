package types

import (
	"github.com/ethereum/go-ethereum/common"
)

/*
BtcAddress is used as an address format that can be validated and marshalled.
Golang's reflection cannot deal with private fields, so (un)marshalling of btcutil.Address does not work.
Therefore, we need this data type for communication.
*/
type EthAddress struct {
	Chain         Chain
	EncodedString string
}

// ParseEthAddress returns an Ethereum address that can be marshalled and checked for correct format.
func ParseEthAddress(address string, chain Chain) (EthAddress, error) {
	addr := EthAddress{EncodedString: address, Chain: chain}
	if err := addr.Validate(); err != nil {
		return EthAddress{}, err
	}
	return addr, nil
}

// Validate does a simple format check
func (a EthAddress) Validate() error {
	if err := a.Chain.Validate(); err != nil {
		return err
	}

	return nil
}

// String returns the encoded address string
func (a EthAddress) String() string {
	return a.EncodedString
}

// Convert decodes the address into a btcutil.Address
func (a EthAddress) Convert() common.Address {
	return common.HexToAddress(a.EncodedString)
}

// PkScript creates a script to pay a transaction output to the address
/*func (a EthAddress) PkScript() []byte {
	addr := a.Convert()

	if script, err := txscript.PayToAddrScript(addr); err != nil {
		return nil
	} else {
		return script
	}
}
*/
