package types

// CLI query error message formats
const (
	ErrFDepositAddr       = "could not resolve deposit address"
	ErrFConsolidationAddr = "could not resolve consolidation address"
	ErrFNextKeyID         = "could not resolve the next key ID"
	ErrFMinOutputAmount   = "could not resolve the minimum output amount allowed"
	ErrFLatestTx          = "could not resolve the latest consolidation transaction"
	ErrFSignedTx          = "could not resolve the signed consolidation transaction"
)
