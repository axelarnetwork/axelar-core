package types

// Event types
const (
	EventTypeNewChain                      = "newChain"
	EventTypeChainConfirmation             = "chainConfirmation"
	EventTypeDepositConfirmation           = "depositConfirmation"
	EventTypeTokenConfirmation             = "tokenConfirmation"
	EventTypeTransferOwnershipConfirmation = "transferOwnershipConfirmation"
)

// Event attribute keys
const (
	AttributeKeyCommandID      = "commandID"
	AttributeKeyChain          = "chain"
	AttributeKeyAddress        = "address"
	AttributeKeyPoll           = "poll"
	AttributeKeyTxID           = "txID"
	AttributeKeyAmount         = "amount"
	AttributeKeyBurnAddress    = "burnAddress"
	AttributeKeyTokenAddress   = "tokenAddress"
	AttributeKeyGatewayAddress = "gatewayAddress"
	AttributeKeyConfHeight     = "confHeight"
	AttributeKeyAsset          = "asset"
	AttributeKeySymbol         = "symbol"
	AttributeKeyNativeAsset    = "nativeAsset"
)

// Event attribute values
const (
	AttributeValueUpdate  = "update"
	AttributeValueStart   = "start"
	AttributeValueReject  = "reject"
	AttributeValueConfirm = "confirm"
)
