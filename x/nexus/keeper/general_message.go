package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getMessageKey(id exported.MessageID) key.Key {
	return generalMessagePrefix.
		Append(key.FromStr(id.ID))
}

func getSentMessageKey(id exported.MessageID) key.Key {
	return sentMessagePrefix.Append(key.From(id.Chain)).Append(key.FromStr(id.ID))
}

// GenerateMessageID generates a unique general message ID
func (k Keeper) GenerateMessageID(ctx sdk.Context, sourceTxID string) string {
	counter := utils.NewCounter[uint](messageNonceKey, k.getStore(ctx))
	nonce := counter.Incr(ctx)

	id := fmt.Sprintf("%s-%d", sourceTxID, nonce)
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

	if m.Is(exported.Sent) {
		if err := k.setSentMessage(ctx, m); err != nil {
			return err
		}
	}

	return k.setMessage(ctx, m)
}

/*
 * Below are the valid message status transitions:
 * Approved -> Sent
 * Sent -> Executed
 * Sent -> Failed
 * Failed -> Sent
 */

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

	if err := k.setMessage(ctx, m); err != nil {
		return err
	}

	return k.setSentMessage(ctx, m)
}

// SetMessageExecuted sets the general message as executed
func (k Keeper) SetMessageExecuted(ctx sdk.Context, messageID exported.MessageID) error {
	m, found := k.GetMessage(ctx, messageID)
	if !found {
		return fmt.Errorf("general message %s not found", messageID.String())
	}

	if !(m.Is(exported.Sent)) {
		return fmt.Errorf("general message is not sent or approved")
	}

	if m.Is(exported.Sent) {
		k.deleteSentMessage(ctx, m)
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

	if !(m.Is(exported.Sent)) {
		return fmt.Errorf("general message is not sent")
	}

	if m.Is(exported.Sent) {
		k.deleteSentMessage(ctx, m)
	}

	m.Status = exported.Failed

	return k.setMessage(ctx, m)
}

// DeleteMessage deletes the general message with associated ID, and also deletes the message ID from the approved messages store
func (k Keeper) DeleteMessage(ctx sdk.Context, messageID exported.MessageID) {
	k.getStore(ctx).DeleteNew(getSentMessageKey(messageID))
	k.getStore(ctx).DeleteNew(getMessageKey(messageID))
}

// GetMessage returns the general message by ID
func (k Keeper) GetMessage(ctx sdk.Context, messageID exported.MessageID) (m exported.GeneralMessage, found bool) {
	return m, k.getStore(ctx).GetNew(getMessageKey(messageID), &m)
}

func (k Keeper) setMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getMessageKey(m.ID), &m)
}

func (k Keeper) setSentMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	if !m.Is(exported.Sent) {
		return fmt.Errorf("general message is not sent")
	}

	return k.getStore(ctx).SetNewValidated(getSentMessageKey(m.ID), &m.ID)
}

func (k Keeper) deleteSentMessage(ctx sdk.Context, m exported.GeneralMessage) {
	k.getStore(ctx).DeleteNew(getSentMessageKey(m.ID))
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

// GetSentMessages returns up to limit sent messages where chain is the destination chain
func (k Keeper) GetSentMessages(ctx sdk.Context, chain exported.ChainName, limit int64) []exported.GeneralMessage {
	msgs := []exported.GeneralMessage{}
	ids := []exported.MessageID{}

	pageRequest := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      uint64(limit),
		CountTotal: false,
		Reverse:    false,
	}

	query.Paginate(prefix.NewStore(k.getStore(ctx).KVStore, append(sentMessagePrefix.Append(key.From(chain)).Bytes(), []byte(key.DefaultDelimiter)...)), pageRequest, func(key []byte, value []byte) error {
		var id exported.MessageID
		k.cdc.MustUnmarshalLengthPrefixed(value, &id)

		ids = append(ids, id)
		return nil
	})

	for _, id := range ids {
		msg, _ := k.GetMessage(ctx, id)
		msgs = append(msgs, msg)
	}

	return msgs
}
