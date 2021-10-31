package types

// Event types
const (
	EventTypeKeygen    = "keygen"
	EventTypeSign      = "sign"
	EventTypeHeartBeat = "heartbeat"
	EventTypeKey       = "key"
)

// Event attribute keys
const (
	AttributeKeyPoll = "poll"
	AttributeChain   = "chain"

	AttributeKeyKeyID     = "keyID"
	AttributeKeyKeyType   = "keyType"
	AttributeKeySigID     = "sigID"
	AttributeKeySigModule = "sigModule"
	AttributeKeySigData   = "sigData"

	AttributeKeySessionID                 = "sessionID"
	AttributeKeyThreshold                 = "threshold"
	AttributeKeyParticipants              = "participants"
	AttributeKeyParticipantShareCounts    = "participantShareCounts"
	AttributeKeyNonParticipants           = "nonParticipants"
	AttributeKeyNonParticipantShareCounts = "nonParticipantShareCounts"
	AttributeKeyPayload                   = "payload"
	AttributeKeyTimeout                   = "timeout"
	AttributeKeyRole                      = "keyRole"
	AttributeKeyKeyIDs                    = "keyIDs"
	AttributeKeyKeyInfos                  = "keyInfos"
)

// Event attribute values
const (
	AttributeValueSend     = "send"
	AttributeValueStart    = "start"
	AttributeValueMsg      = "message"
	AttributeValueDecided  = "decided"
	AttributeValueReject   = "reject"
	AttributeValueAssigned = "assigned"
)
