package types

// EventTypeOutpointConfirmation is an event type
const (
	EventTypeOutpointConfirmation = "outpointConfirmation"
)

// Event attribute keys
const (
	AttributeKeyConfHeight   = "confHeight"
	AttributeKeyOutPointInfo = "outPointInfo"
	AttributeKeyPoll         = "poll"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueConfirm = "confirm"
	AttributeValueReject  = "reject"
)
