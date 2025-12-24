package types

// Event types
const (
	EventTypeChain           = "chain"
	EventTypeChainMaintainer = "chainMaintainer"
)

// Event attribute keys
const (
	AttributeKeyChain                  = "chain"
	AttributeKeyChainMaintainerAddress = "chainMaintainerAddress"
	AttributeKeyAsset                  = "asset"
	AttributeKeyBlock                  = "block"
	AttributeKeyMessageID              = "messageID"
	AttributeKeySourceChain            = "sourceChain"
	AttributeKeyDestinationChain       = "destinationChain"
)

// Event attribute values
const (
	AttributeValueRegister    = "register"
	AttributeValueDeregister  = "deregister"
	AttributeValueActivated   = "activated"
	AttributeValueDeactivated = "deactivated"
)
