package types

// CLI query error message formats
const (
	ErrDepositAddr       = "could not resolve deposit address"
	ErrDepositStatus     = "could not resolve deposit status"
	ErrConsolidationAddr = "could not resolve consolidation address"
	ErrNextKeyID         = "could not resolve the next key ID"
	ErrExternalKeyID     = "could not resolve the external key IDs"
	ErrMinOutputAmount   = "could not resolve the minimum output amount allowed"
	ErrLatestTx          = "could not resolve the latest consolidation transaction"
	ErrSignedTx          = "could not resolve the signed consolidation transaction"
)
