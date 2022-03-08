package types

// Event types
const (
	EventTypeDepositConfirmation = "depositConfirmation"
	EventTypeLink                = "link"
)

// Event attribute keys
const (
	AttributeKeyChain              = "chain"
	AttributeKeySourceChain        = "sourceChain"
	AttributeKeyTxID               = "txID"
	AttributeKeyAsset              = "asset"
	AttributeKeyDepositAddress     = "depositAddress"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyDestinationChain   = "destinationChain"
	AttributeKeyTransferID         = "transferID"
)

// Event attribute values
const (
	AttributeValueConfirm = "confirm"
)
