package types

// EventTypeOutpointConfirmation is an event type
const (
	EventTypeOutpointConfirmation = "outpointConfirmation"
	EventTypeTransactionSigned    = "transactionSigned"
	EventTypeWithdrawal           = "withdrawal"
)

// Event attribute keys
const (
	AttributeKeyConfHeight         = "confHeight"
	AttributeKeyOutPointInfo       = "outPointInfo"
	AttributeKeyPoll               = "poll"
	AttributeKeyTxHash             = "txHash"
	AttributeKeyAmount             = "amount"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyDestinationChain   = "destinationChain"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueConfirm = "confirm"
	AttributeValueReject  = "reject"
	AttributeValueFailed  = "failed"
)
