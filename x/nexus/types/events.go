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
	AttributeKeyMessageId              = "messageId"
	AttributeKeyBlock                  = "block"
	AttributeKeyIsBeginEndBlocker      = "isBeginOrEndBlocker"
	AttributeKeyTxHash                 = "txHash"
)

// Event attribute values
const (
	AttributeValueRegister    = "register"
	AttributeValueDeregister  = "deregister"
	AttributeValueActivated   = "activated"
	AttributeValueDeactivated = "deactivated"
)
