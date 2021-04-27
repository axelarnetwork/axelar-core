package types

// Event types
const (
	EventTypeKeygen = "keygen"
	EventTypeSign   = "sign"
)

// Event attribute keys
const (
	AttributeKeyPoll = "poll"
	AttributeChain   = "chain"

	AttributeKeyKeyID                  = "keyID"
	AttributeKeySigID                  = "sigID"
	AttributeKeySessionID              = "sessionID"
	AttributeKeyThreshold              = "threshold"
	AttributeKeyParticipants           = "participants"
	AttributeKeyParticipantShareCounts = "participantShareCounts"
	AttributeKeyPayload                = "payload"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueMsg     = "message"
	AttributeValueDecided = "decided"
)
