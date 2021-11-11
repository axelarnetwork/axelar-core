package types

// Event types
const (
	EventTypeNewChain                      = "newChain"
	EventTypeGatewayDeploymentConfirmation = "gatewayDeploymentConfirmation"
	EventTypeChainConfirmation             = "chainConfirmation"
	EventTypeDepositConfirmation           = "depositConfirmation"
	EventTypeTokenConfirmation             = "tokenConfirmation"
	EventTypeTransferKeyConfirmation       = "transferKeyConfirmation"
	EventTypeLink                          = "link"
)

// Event attribute keys
const (
	AttributeKeyBatchedCommandsID  = "batchedCommandID"
	AttributeKeyChain              = "chain"
	AttributeKeyAddress            = "address"
	AttributeKeyThreshold          = "threshold"
	AttributeKeyPoll               = "poll"
	AttributeKeyTxID               = "txID"
	AttributeKeyTransferKeyType    = "transferKeyType"
	AttributeKeyKeyType            = "keyType"
	AttributeKeyAmount             = "amount"
	AttributeKeyBurnAddress        = "burnAddress"
	AttributeKeyTokenAddress       = "tokenAddress"
	AttributeKeyGatewayAddress     = "gatewayAddress"
	AttributeKeyBytecodeHash       = "bytecodeHash"
	AttributeKeyConfHeight         = "confHeight"
	AttributeKeyAsset              = "asset"
	AttributeKeySymbol             = "symbol"
	AttributeKeyNativeAsset        = "nativeAsset"
	AttributeKeyDestinationChain   = "destinationChain"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyValue              = "value"
)

// Event attribute values
const (
	AttributeValueUpdate  = "update"
	AttributeValueStart   = "start"
	AttributeValueReject  = "reject"
	AttributeValueConfirm = "confirm"
	AttributeValueVote    = "vote"
)
