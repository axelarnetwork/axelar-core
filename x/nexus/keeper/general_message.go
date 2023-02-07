package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getMessageKey(id exported.MessageID) key.Key {
	return generalMessagePrefix.
		Append(key.FromStr(id.ID))
}

func getApprovedMessageKey(id exported.MessageID) key.Key {
	return approvedGeneralMessagePrefix.Append(key.From(id.Chain)).Append(key.FromStr(id.ID))
}

// GetGeneralMessageID generates a unique general message ID
func (k Keeper) GetGeneralMessageID(ctx sdk.Context, sourceTxID string) string {
	counter := utils.NewCounter[uint](messageNonceKey, k.getStore(ctx))
	nonce := counter.Incr(ctx)

	id := fmt.Sprintf("%s-%x", sourceTxID, nonce)
	return id
}

// SetNewMessage sets the given general message. If the messages is approved, adds the message ID to approved messages store
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

	if _, found := k.GetMessage(ctx, m.ID); found {
		return fmt.Errorf("general message %s already exists", m.ID)
	}
	if m.Is(exported.Approved) {
		if err := k.setApprovedMessage(ctx, m); err != nil {
			return err
		}
	}
	return k.setMessage(ctx, m)
}

// SetMessageApproved sets the message as approved, and adds the message ID to the approved messages store
func (k Keeper) SetMessageApproved(ctx sdk.Context, messageID exported.MessageID) error {

	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	m.Status = exported.Approved
	if err := k.setMessage(ctx, m); err != nil {
		return err
	}
	return k.setApprovedMessage(ctx, m)
}

// SetMessageSent sets the general message as sent
func (k Keeper) SetMessageSent(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Approved) || m.Is(exported.Failed)) {
		return fmt.Errorf("general message is not approved or failed")
	}
	if m.Is(exported.Approved) {
		k.deleteApprovedMessage(ctx, m)
	}

	m.Status = exported.Sent

	return k.setMessage(ctx, m)
}

// SetMessageExecuted sets the general message as executed
func (k Keeper) SetMessageExecuted(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Sent) || m.Is(exported.Approved)) {
		return fmt.Errorf("general message is not sent or approved")
	}
	if m.Is(exported.Approved) {
		k.deleteApprovedMessage(ctx, m)
	}

	m.Status = exported.Executed

	return k.setMessage(ctx, m)
}

// SetMessageFailed sets the general message as failed
func (k Keeper) SetMessageFailed(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Sent) || m.Is(exported.Approved)) {
		return fmt.Errorf("general message is not sent or approved")
	}
	if m.Is(exported.Approved) {
		k.deleteApprovedMessage(ctx, m)
	}

	m.Status = exported.Failed

	return k.setMessage(ctx, m)
}

// DeleteMessage deletes the general message with associated ID, and also deletes the message ID from the approved messages store
func (k Keeper) DeleteMessage(ctx sdk.Context, messageID exported.MessageID) {
	k.getStore(ctx).DeleteNew(getApprovedMessageKey(messageID))
	k.getStore(ctx).DeleteNew(getMessageKey(messageID))
}

// GetMessage returns the general message by ID
func (k Keeper) GetMessage(ctx sdk.Context, messageID exported.MessageID) (m exported.GeneralMessage, found bool) {
	return m, k.getStore(ctx).GetNew(getMessageKey(messageID), &m)
}

func (k Keeper) setMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getMessageKey(m.ID), &m)
}

func (k Keeper) setApprovedMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getApprovedMessageKey(m.ID), &m.ID)
}

func (k Keeper) deleteApprovedMessage(ctx sdk.Context, m exported.GeneralMessage) {
	k.getStore(ctx).DeleteNew(getApprovedMessageKey(m.ID))
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

// GetApprovedMessages returns up to limit approved messages where chain is the destination chain
func (k Keeper) GetApprovedMessages(ctx sdk.Context, chain exported.ChainName, limit int64) []exported.GeneralMessage {
	msgs := []exported.GeneralMessage{}
	ids := []exported.MessageID{}
	iter := k.getStore(ctx).IteratorNew(approvedGeneralMessagePrefix.Append(key.From(chain)))
	defer utils.CloseLogError(iter, k.Logger(ctx))
	for i := 0; iter.Valid() && i < int(limit); iter.Next() {
		var id exported.MessageID
		iter.UnmarshalValue(&id)
		ids = append(ids, id)
		i++
	}
	for _, id := range ids {
		msg, _ := k.GetMessage(ctx, id)
		msgs = append(msgs, msg)
	}
	return msgs
}
