package types

// Event types
const (
	EventTypeDepositConfirmation = "depositConfirmation"
	EventTypeTokenConfirmation   = "tokenConfirmation"
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
	AttributeKeySymbol         = "symbol"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueReject  = "reject"
	AttributeValueConfirm = "confirm"
)
