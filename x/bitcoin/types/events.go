package types

// EventTypeOutpointConfirmation is an event type
const (
	EventTypeExternalSignature    = "externalSignature"
	EventTypeKey                  = "key"
	EventTypeConsolidationTx      = "consolidationTransaction"
	EventTypeOutpointConfirmation = "outpointConfirmation"
	EventTypeWithdrawal           = "withdrawal"
)

// Event attribute keys
const (
	AttributeKeyKeyID              = "keyID"
	AttributeKeyRole               = "keyRole"
	AttributeKeySigID              = "sigID"
	AttributeKeyConfHeight         = "confHeight"
	AttributeKeyOutPointInfo       = "outPointInfo"
	AttributeKeyPoll               = "poll"
	AttributeKeyAmount             = "amount"
	AttributeKeyMasterKeyID        = "masterKeyId"
	AttributeKeySecondaryKeyID     = "secondaryKeyId"
	AttributeKeyDepositAddress     = "depositAddress"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyDestinationChain   = "destinationChain"
)

// Event attribute values
const (
	AttributeValueSubmitted      = "submitted"
	AttributeValueAssigned       = "assigned"
	AttributeValueCreated        = "created"
	AttributeValueSigning        = "signing"
	AttributeValueSigningAborted = "signingAborted"
	AttributeValueSigned         = "signed"
	AttributeValueStart          = "start"
	AttributeValueConfirm        = "confirm"
	AttributeValueReject         = "reject"
	AttributeValueFailed         = "failed"
)
