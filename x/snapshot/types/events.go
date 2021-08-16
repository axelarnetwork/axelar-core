package types

// Event types
const (
	EventTypeCreateSnapshot = "createSnapshot"
)

// Event attribute keys
const (
	AttributeModule         		= ModuleName
	AttributeAddress        		= "address"
	AttributeRegisterProxy  		= "registerProxy"
	AttributeDeactivateProxy 		= "deactivateProxy"
	AttributeParticipants      		= "participants"
	AttributeParticipantsStake 		= "participantsStake"
	AttributeNonParticipants      	= "nonParticipants"
	AttributeNonParticipantsStake 	= "nonParticipantsStake"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
)
