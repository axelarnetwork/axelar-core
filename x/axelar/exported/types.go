package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExternalChainAddress struct {
	Chain   string
	Address string
}

func (addr ExternalChainAddress) IsInvalid() bool {
	return addr.Chain == "" || addr.Address == ""
}

func (addr ExternalChainAddress) String() string {
	return fmt.Sprintf("chain: %s, address: %s", addr.Chain, addr.Address)
}

type ExternalTx struct {
	Chain  string
	TxID   string
	Amount sdk.DecCoin
}

func (tx ExternalTx) IsInvalid() bool {
	return tx.Chain == "" ||
		tx.TxID == "" ||
		!tx.Amount.IsValid() ||
		!tx.Amount.IsPositive()
}

func (tx ExternalTx) String() string {
	return fmt.Sprintf("chain: %s, txID: %s, amount: %s", tx.Chain, tx.TxID, tx.Amount.String())
}
