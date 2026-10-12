package types

// event types and attributes
const (
	EventTypeEmergencyPause   = "emergency_pause"
	EventTypeEmergencyUnpause = "emergency_unpause"
	EventTypeEmergencyHalt    = "emergency_halt"

	AttributeKeySigner             = "signer"
	AttributeKeyReason             = "reason"
	AttributeKeyHeight             = "height"
	AttributeKeyPauseLevel         = "pause_level"
	AttributeKeyPreviousPauseLevel = "previous_pause_level"
	AttributeKeyHaltHeight         = "halt_height"
	AttributeKeyPlanName           = "plan_name"
	AttributeKeyReplacedPlanName   = "replaced_plan_name"
	AttributeKeyReplacedPlanHeight = "replaced_plan_height"
	AttributeKeyReplacedPlanInfo   = "replaced_plan_info"
)
