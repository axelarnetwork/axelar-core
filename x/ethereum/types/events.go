package types

// Event types
const (
	EventTypeDepositConfirmation = "depositConfirmation"
	EventTypeTokenConfirmation   = "tokenConfirmation"
)

// Attributes
const (
	AttributeModule    = ModuleName
	AttributeAddress   = "address"
	AttributeCommandID = "commandID"
)

// Event attribute keys
const (
	AttributeKeyPoll           = "poll"
	AttributeKeyResult         = "result"
	AttributeKeyTxID           = "txID"
	AttributeKeyAmount         = "amount"
	AttributeKeyBurnAddress    = "burnAddress"
	AttributeKeyTokenAddress   = "tokenAddress"
	AttributeKeyGatewayAddress = "gatewayAddress"
	AttributeKeyConfHeight     = "confHeight"
	AttributeKeySymbol         = "symbol"
	AttributeKeyDeploySig      = "deploySig"
)

// Event attribute values
const (
	AttributeKeyActionToken   = "tokenVerify"
	AttributeKeyActionDeposit = "depositVerify"
	AttributeKeyActionUnknown = "unknownVerify"
	AttributeValueStart       = "start"
)
