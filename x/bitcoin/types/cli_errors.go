package types

// CLI query error message formats
const (
	ErrFDepositAddress = "could not resolve master key"
	ErrFTxInfo         = "could not resolve txID %s and vout index %d"
	ErrFSendTransfers  = "could not send the consolidation transaction"
)
