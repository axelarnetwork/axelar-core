package types

// CLI query error message formats
const (
	ErrFDepositAddress       = "could not resolve master key"
	ErrFGetTransfers         = "could not get the consolidation transaction"
	ErrFGetSignTransferState = "could not get the sign transfer state"
	ErrFMinWithdraw			 = "could not get the minimum withdraw amount"
	ErrFTxState				 = "could not get bitcoin transaction state"
	ErrFConsolidationState	 = "could not get bitcoin consolidation transaction state"
)
