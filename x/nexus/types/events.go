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
	AttributeKeyLimit                  = "limit"
	AttributeKeyTransferEpoch          = "transferEpoch"
	AttributeKeyBlock                  = "block"
)

// Event attribute values
const (
	AttributeValueRegister    = "register"
	AttributeValueDeregister  = "deregister"
	AttributeValueActivated   = "activated"
	AttributeValueDeactivated = "deactivated"
)
