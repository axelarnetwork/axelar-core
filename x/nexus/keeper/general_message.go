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
		Append(key.From(id.Chain)).
		Append(key.FromStr(id.ID))
}

// SetNewMessage sets the given general message
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

	return k.setMessage(ctx, m)
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

	m.Status = exported.Sent

	return k.setMessage(ctx, m)
}

// SetMessageExecuted sets the general message as executed
func (k Keeper) SetMessageExecuted(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessage(ctx, messageID)
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
	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !m.Is(exported.Sent) {
		return fmt.Errorf("general message is not sent")
	}

	m.Status = exported.Failed

	return k.setMessage(ctx, m)
}

// GetMessage returns the general message by ID
func (k Keeper) GetMessage(ctx sdk.Context, messageID exported.MessageID) (m exported.GeneralMessage, found bool) {
	return m, k.getStore(ctx).GetNew(getMessageKey(messageID), &m)
}

// DeleteMessage returns the general message by ID
func (k Keeper) DeleteMessage(ctx sdk.Context, messageID exported.MessageID) {
	k.getStore(ctx).DeleteNew(getMessageKey(messageID))
}

func (k Keeper) setMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getMessageKey(m.ID), &m)
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
