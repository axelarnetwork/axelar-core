package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getMessageKey(id exported.MessageID, status exported.GeneralMessage_Status) key.Key {
	return generalMessagePrefix.
		Append(key.From(id.Chain)).
		Append(key.From(status)).
		Append(key.FromStr(id.ID))
}

// GetGeneralMessageID generates a unique general message ID
func (k Keeper) GetGeneralMessageID(ctx sdk.Context, sourceTxID string, sourceChain exported.ChainName) string {
	counter := utils.NewCounter[uint](generalMessageNonceKey, k.getStore(ctx))
	nonce := counter.Incr(ctx)

	id := fmt.Sprintf("%s-%s-%x", sourceTxID, sourceChain, nonce)
	return id
}

// SetNewMessage sets the given general message and enqueues the message for the destination chain
func (k Keeper) SetNewMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	sourceChain, ok := k.GetChain(ctx, m.SourceChain)
	if !ok {
		return fmt.Errorf("source chain %s is not a registered chain", m.SourceChain)
	}

	destChain, ok := k.GetChain(ctx, m.ID.Chain)
	if !ok {
		return fmt.Errorf("destination chain %s is not a registered chain", m.ID.Chain)
	}

	validator := k.GetRouter().GetAddressValidator(destChain.Module)
	if err := validator(ctx, exported.CrossChainAddress{Chain: destChain, Address: m.Receiver}); err != nil {
		return err
	}

	if m.Asset != nil {
		if err := k.validateTransferAsset(ctx, sourceChain, m.Asset.Denom); err != nil {
			return err
		}

		if err := k.validateTransferAsset(ctx, destChain, m.Asset.Denom); err != nil {
			return err
		}
	}

	if _, found := k.GetMessageAnyStatus(ctx, m.ID); found {
		return fmt.Errorf("general message %s already exists", m.ID)
	}
	return k.setMessage(ctx, m)
}

// SetMessageSent sets the general message as sent
func (k Keeper) SetMessageSent(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessageAnyStatus(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Approved) || m.Is(exported.Failed)) {
		return fmt.Errorf("general message is not approved or failed")
	}

	m.Status = exported.Sent

	return k.setMessage(ctx, m)
}

// SetMessageExecuted sets the general message as executed
func (k Keeper) SetMessageExecuted(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessageAnyStatus(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !m.Is(exported.Sent) {
		return fmt.Errorf("general message is not sent")
	}

	m.Status = exported.Executed

	return k.setMessage(ctx, m)
}

// SetMessageFailed sets the general message as failed
func (k Keeper) SetMessageFailed(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessageAnyStatus(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Sent) || m.Is(exported.Approved)) {
		return fmt.Errorf("general message is not sent or approved")
	}

	m.Status = exported.Failed

	return k.setMessage(ctx, m)
}

// GetMessageAnyStatus returns the general message matching the given ID
func (k Keeper) GetMessageAnyStatus(ctx sdk.Context, messageID exported.MessageID) (m exported.GeneralMessage, found bool) {
	return k.GetMessageWithStatus(ctx, messageID, []exported.GeneralMessage_Status{})
}

// GetMessageWithStatus returns the general message matching the given ID and one of the passed in statuses
func (k Keeper) GetMessageWithStatus(ctx sdk.Context, messageID exported.MessageID, statuses []exported.GeneralMessage_Status) (m exported.GeneralMessage, found bool) {
	if len(statuses) == 0 {
		statuses = []exported.GeneralMessage_Status{exported.Sent, exported.Approved, exported.Executed, exported.Failed}
	}
	for _, s := range statuses {
		if found := k.getStore(ctx).GetNew(getMessageKey(messageID, s), &m); found {
			return m, found
		}
	}
	return m, false
}

// DeleteMessage returns the general message by ID
func (k Keeper) DeleteMessage(ctx sdk.Context, messageID exported.MessageID) {
	statuses := []exported.GeneralMessage_Status{exported.Sent, exported.Approved, exported.Executed, exported.Failed}
	for _, s := range statuses {
		k.getStore(ctx).DeleteNew(getMessageKey(messageID, s))
	}
}

func (k Keeper) setMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getMessageKey(m.ID, m.Status), &m)
}

func (k Keeper) getMessages(ctx sdk.Context) (generalMessages []exported.GeneralMessage) {
	iter := k.getStore(ctx).IteratorNew(generalMessagePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var generalMessage exported.GeneralMessage
		iter.UnmarshalValue(&generalMessage)

		generalMessages = append(generalMessages, generalMessage)
	}

	return generalMessages
}

// ConsumeApprovedMessages returns up to limit messages where chain is the destination chain, and deletes the messages from the store
func (k Keeper) ConsumeApprovedMessages(ctx sdk.Context, chain exported.ChainName, limit int64) []exported.GeneralMessage {
	msgs := []exported.GeneralMessage{}
	iter := k.getStore(ctx).IteratorNew(generalMessagePrefix.Append(key.From(chain)).Append(key.From(exported.Approved)))
	defer utils.CloseLogError(iter, k.Logger(ctx))
	for i := 0; iter.Valid() && i < int(limit); iter.Next() {
		var msg exported.GeneralMessage
		iter.UnmarshalValue(&msg)
		msgs = append(msgs, msg)
		i++
		k.getStore(ctx).DeleteRaw(iter.Key())
	}
	return msgs
}
