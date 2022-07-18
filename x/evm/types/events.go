package types

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Event types
const (
	EventTypeNewChain                = "newChain"
	EventTypeGateway                 = "gateway"
	EventTypeChainConfirmation       = "chainConfirmation"
	EventTypeGatewayTxConfirmation   = "gatewayTxConfirmation"
	EventTypeDepositConfirmation     = "depositConfirmation"
	EventTypeTokenConfirmation       = "tokenConfirmation"
	EventTypeTransferKeyConfirmation = "transferKeyConfirmation"
	EventTypeLink                    = "link"
	EventTypeSign                    = "sign"
	EventTypeEventConfirmation       = "eventConfirmation"
)

// Event attribute keys
const (
	AttributeKeyBatchedCommandsID  = "batchedCommandID"
	AttributeKeyChain              = "chain"
	AttributeKeySourceChain        = "sourceChain"
	AttributeKeyAddress            = "address"
	AttributeKeyThreshold          = "threshold"
	AttributeKeyPoll               = "poll"
	AttributeKeyTxID               = "txID"
	AttributeKeyKeyType            = "keyType"
	AttributeKeyAmount             = "amount"
	AttributeKeyDepositAddress     = "depositAddress"
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
	AttributeKeyCommandsID         = "commandID"
	AttributeKeyCommandsIDs        = "commandIDs"
	AttributeKeyTransferID         = "transferID"
	AttributeKeyEventType          = "eventType"
	AttributeKeyEventID            = "eventID"
	AttributeKeyKeyID              = "keyID"
)

// Event attribute values
const (
	AttributeValueUpdate  = "update"
	AttributeValueStart   = "start"
	AttributeValueReject  = "reject"
	AttributeValueConfirm = "confirm"
	AttributeValueVote    = "vote"
)

// NewConfirmKeyTransfer is the constructor for event confirm key transfer
func NewConfirmKeyTransfer(chain nexus.ChainName, txID Hash, gatewayAddress Address, confirmationHeight uint64, pollID vote.PollID) *ConfirmKeyTransfer {
	return &ConfirmKeyTransfer{
		Chain:              chain,
		TxID:               txID,
		GatewayAddress:     gatewayAddress,
		ConfirmationHeight: confirmationHeight,
		PollID:             pollID,
	}
}
