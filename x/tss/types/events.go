package types

// Event types
const (
	EventTypeKeygen = "keygen"
	EventTypeSign   = "sign"
	EventTypeAck    = "ack"
)

// Event attribute keys
const (
	AttributeKeyPoll = "poll"
	AttributeChain   = "chain"

	AttributeKeyKeyID                  = "keyID"
	AttributeKeySigID                  = "sigID"
	AttributeKeyHeight                 = "height"
	AttributeKeyKeyAckType             = "ackType"
	AttributeKeySessionID              = "sessionID"
	AttributeKeyThreshold              = "threshold"
	AttributeKeyParticipants           = "participants"
	AttributeKeyParticipantShareCounts = "participantShareCounts"
	AttributeKeyPayload                = "payload"
	AttributeKeyTimeout                = "timeout"
)

// Event attribute values
const (
	AttributeValueKeygen  = "keygen"
	AttributeValueSign    = "sign"
	AttributeValueStart   = "start"
	AttributeValueMsg     = "message"
	AttributeValueDecided = "decided"
	AttributeValueReject  = "reject"
)
