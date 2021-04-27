package types

// EventTypeOutpointConfirmation is an event type
const (
	EventTypeOutpointConfirmation = "outpointConfirmation"
	EventTypeTransactionSigned    = "transactionSigned"
	EventTypeWithdrawalFailed     = "withdrawalFailed"
)

// Event attribute keys
const (
	AttributeKeyConfHeight         = "confHeight"
	AttributeKeyOutPointInfo       = "outPointInfo"
	AttributeKeyPoll               = "poll"
	AttributeKeyTxHash             = "txHash"
	AttributeKeyAmount             = "amount"
	AttributeKeyDestinationAddress = "destinationAddress"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueConfirm = "confirm"
	AttributeValueReject  = "reject"
)
