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
	AttributeKeySourceAddress      = "sourceAddress"
	AttributeKeyTxID               = "txID"
	AttributeKeyAsset              = "asset"
	AttributeKeyDepositAddress     = "depositAddress"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyDestinationChain   = "destinationChain"
	AttributeKeyTransferID         = "transferID"
	AttributeKeyPayloadHash        = "payloadHash"
	AttributeKeyMessageID          = "messageID"
)

// Event attribute values
const (
	AttributeValueConfirm = "confirm"
)
