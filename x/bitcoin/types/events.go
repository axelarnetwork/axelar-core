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
	AttributeKeyDestinationAddress = "destinationAddress"
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
