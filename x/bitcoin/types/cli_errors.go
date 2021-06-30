package types

// CLI query error message formats
const (
	ErrFMasterKey            = "could not resolve master key"
	ErrFGetRawTx             = "could not get the raw consolidation transaction"
	ErrFGetPayForRawTx       = "could not get the raw pay-for-consolidation transaction"
	ErrFInvalidFeeRate       = "invalid fee rate"
	ErrFGetSignTransferState = "could not get the sign transfer state"
	ErrFMinWithdraw          = "could not get the minimum withdraw amount"
	ErrFTxState              = "could not get bitcoin transaction state"
	ErrFConsolidationState   = "could not get bitcoin consolidation transaction state"
)
