package types

// Event types
const (
	EventTypeKeygen = "keygen"
	EventTypeSign   = "sign"
	EventTypeAck    = "ack"
	EventTypeKey    = "key"
)

// Event attribute keys
const (
	AttributeKeyPoll = "poll"
	AttributeChain   = "chain"

	AttributeKeyKeyID     = "keyID"
	AttributeKeySigID     = "sigID"
	AttributeKeySigModule = "sigModule"
	AttributeKeySigData   = "sigData"

	AttributeKeyHeight                    = "height"
	AttributeKeyKeyAckType                = "ackType"
	AttributeKeySessionID                 = "sessionID"
	AttributeKeyThreshold                 = "threshold"
	AttributeKeyParticipants              = "participants"
	AttributeKeyParticipantShareCounts    = "participantShareCounts"
	AttributeKeyNonParticipants           = "nonParticipants"
	AttributeKeyNonParticipantShareCounts = "nonParticipantShareCounts"
	AttributeKeyPayload                   = "payload"
	AttributeKeyTimeout                   = "timeout"
	AttributeKeyDidStart                  = "didStart"
	AttributeKeyRole                      = "keyRole"
)

// Event attribute values
const (
	AttributeValueKeygen   = "keygen"
	AttributeValueSign     = "sign"
	AttributeValueStart    = "start"
	AttributeValueMsg      = "message"
	AttributeValueDecided  = "decided"
	AttributeValueReject   = "reject"
	AttributeValueAssigned = "assigned"
)
