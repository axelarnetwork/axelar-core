package types

// Event types
const (
	EventTypeDepositConfirmation = "depositConfirmation"
	EventTypeLink                = "link"
	EventTypeCallContract        = "callContract"
)

// Event attribute keys
const (
	AttributeKeyChain               = "chain"
	AttributeKeySourceChain         = "sourceChain"
	AttributeKeySourceAddress       = "sourceAddress"
	AttributeKeyTxID                = "txID"
	AttributeKeyAsset               = "asset"
	AttributeKeyDepositAddress      = "depositAddress"
	AttributeKeyDestinationAddress  = "destinationAddress"
	AttributeKeyDestinationChain    = "destinationChain"
	AttributeKeyTransferID          = "transferID"
	AttributeKeyContractPayload     = "contractPayload"
	AttributeKeyContractPayloadHash = "contractPayloadHash"
	AttributeKeyCommandID           = "commandID"
)

// Event attribute values
const (
	AttributeValueConfirm = "confirm"
)
