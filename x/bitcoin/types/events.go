package types

// EventTypeOutpointConfirmation is an event type
const (
	EventTypeOutpointConfirmation = "outpointConfirmation"
	EventTypeTransactionSigned    = "transactionSigned"
)

// Event attribute keys
const (
	AttributeKeyConfHeight   = "confHeight"
	AttributeKeyOutPointInfo = "outPointInfo"
	AttributeKeyPoll         = "poll"
	AttributeKeyTxHash       = "txHash"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueConfirm = "confirm"
	AttributeValueReject  = "reject"
)
