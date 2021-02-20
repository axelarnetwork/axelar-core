package types

// CLI query error message formats
const (
	ErrFDepositAddress = "could not resolve master key"
	ErrFTxInfo         = "could not resolve txID %s and vout index %d"
	ErrFSendTx         = "could not send a consolidation transaction"
)
