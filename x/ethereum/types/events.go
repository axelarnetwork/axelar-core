package types

// Attributes
const (
	AttributeModule        = ModuleName
	AttributeAddress       = "address"
	AttributeBurnAddress   = "burnAddress"
	AttributeTxID          = "txID"
	AttributeCommandID     = "commandID"
	AttributeAmount        = "amount"
	AttributeDestination   = "destination"
	AttributeVotingData    = "data"
	AttributePollConfirmed = "confirmed"
)

// EventTypeVerificationResult describes an event type
const (
	EventTypeVerificationResult = "verificationResult"
)

// Event attribute keys
const (
	AttributeKeyPoll   = "poll"
	AttributeKeyResult = "result"
	AttributeKeyTxID   = "txID"
)

// Event attribute values
const (
	AttributeKeyActionToken   = "tokenVerify"
	AttributeKeyActionDeposit = "depositVerify"
	AttributeKeyActionUnknown = "unknownVerify"
)
