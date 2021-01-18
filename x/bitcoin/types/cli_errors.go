package types

// CLI query error message formats
const (
	ErrFDepositAddress       = "could not resolve master key"
	ErrFConsolidationAddress = "could not resolve master key"
	ErrFTxInfo               = "could not resolve txID %s and vout index %d"
	ErrFRawTx                = "could not create a new transaction spending transaction %s"
	ErrFSendTx               = "could not send the transaction spending transaction %s"
)

