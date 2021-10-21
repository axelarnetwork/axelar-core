package types

// Event types
const (
	EventTypeDepositConfirmation = "depositConfirmation"
	EventTypeLink                = "link"
)

// Event attribute keys
const (
	AttributeKeyChain              = "chain"
	AttributeKeyTxID               = "txID"
	AttributeKeyAmount             = "amount"
	AttributeKeyDepositAddress     = "depositAddress"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyDestinationChain   = "destinationChain"
)

// Event attribute values
const (
	AttributeValueConfirm = "confirm"
)
