package types

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

func ValidTxJson(txJson []byte) error {
	var tx *ethTypes.Transaction
	return ModuleCdc.UnmarshalJSON(txJson, &tx)
}

func ValidAddress(address string) bool {
	if bytes.Equal(common.HexToAddress(address).Bytes(), make([]byte, common.AddressLength)) {
		return false
	}

	return true
}
