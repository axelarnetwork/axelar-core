package types

import (
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Event types
const (
	EventTypeNewChain                = "newChain"
	EventTypeGateway                 = "gateway"
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
	AttributeKeyPoll               = "poll"
	AttributeKeyTxID               = "txID"
	AttributeKeyAmount             = "amount"
	AttributeKeyDepositAddress     = "depositAddress"
	AttributeKeyTokenAddress       = "tokenAddress"
	AttributeKeyGatewayAddress     = "gatewayAddress"
	AttributeKeyConfHeight         = "confHeight"
	AttributeKeyAsset              = "asset"
	AttributeKeySymbol             = "symbol"
	AttributeKeyDestinationChain   = "destinationChain"
	AttributeKeyDestinationAddress = "destinationAddress"
	AttributeKeyCommandsID         = "commandID"
	AttributeKeyCommandsIDs        = "commandIDs"
	AttributeKeyTransferID         = "transferID"
	AttributeKeyEventType          = "eventType"
	AttributeKeyEventID            = "eventID"
	AttributeKeyKeyID              = "keyID"
	AttributeKeyMessageID          = "messageID"
)

// Event attribute values
const (
	AttributeValueStart   = "start"
	AttributeValueConfirm = "confirm"
)

// NewConfirmKeyTransferStarted is the constructor for event confirm key transfer
func NewConfirmKeyTransferStarted(chain nexus.ChainName, txID Hash, gatewayAddress Address, confirmationHeight uint64, participants vote.PollParticipants) *ConfirmKeyTransferStarted {
	return &ConfirmKeyTransferStarted{
		Chain:              chain,
		TxID:               txID,
		GatewayAddress:     gatewayAddress,
		ConfirmationHeight: confirmationHeight,
		PollParticipants:   participants,
	}
}

// NewCommandBatchSigned returns a new CommandBatchSigned instance
func NewCommandBatchSigned(chain nexus.ChainName, batchID []byte) *CommandBatchSigned {
	return &CommandBatchSigned{Chain: chain, CommandBatchID: batchID}
}

// NewCommandBatchAborted returns a new CommandBatchAborted instance
func NewCommandBatchAborted(chain nexus.ChainName, batchID []byte) *CommandBatchAborted {
	return &CommandBatchAborted{Chain: chain, CommandBatchID: batchID}
}
