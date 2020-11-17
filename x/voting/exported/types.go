package exported

import (
	"fmt"
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
	Chain string
	TxID  string
}

func (tx ExternalTx) IsInvalid() bool {
	return tx.Chain == "" ||
		tx.TxID == ""
}

func (tx ExternalTx) String() string {
	return fmt.Sprintf(
		"chain: %s, txID: %s",
		tx.Chain,
		tx.TxID,
	)
}

type FutureVote struct {
	Tx          ExternalTx
	LocalAccept bool
}
