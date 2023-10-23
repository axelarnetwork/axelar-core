package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func getMessageKey(id string) key.Key {
	return generalMessagePrefix.Append(key.FromStr(id))
}

func getProcessingMessageKey(destinationChain exported.ChainName, id string) key.Key {
	return processingMessagePrefix.Append(key.From(destinationChain)).Append(key.FromStr(id))
}

// GenerateMessageID generates a unique general message ID, and returns the message ID, current transacation ID and a unique integer nonce
// The message ID is just a concatenation of the transaction ID and the nonce
func (k Keeper) GenerateMessageID(ctx sdk.Context) (string, []byte, uint64) {
	counter := utils.NewCounter[uint64](messageNonceKey, k.getStore(ctx))
	nonce := counter.Incr(ctx)
	hash := sha256.Sum256(ctx.TxBytes())

	return fmt.Sprintf("0x%s-%d", hex.EncodeToString(hash[:]), nonce), hash[:], nonce
}

// SetNewWasmMessage sets the given general message from a wasm contract.
// Deprecated
func (k Keeper) SetNewWasmMessage(ctx sdk.Context, msg exported.GeneralMessage) error {
	if msg.Asset != nil {
		return fmt.Errorf("asset transfer is not supported")
	}

	if _, ok := k.GetChain(ctx, msg.GetDestinationChain()); !ok {
		return fmt.Errorf("destination chain %s is not a registered chain", msg.GetDestinationChain())
	}

	if !k.IsChainActivated(ctx, msg.Recipient.Chain) {
		return fmt.Errorf("destination chain %s is not activated", msg.GetDestinationChain())
	}

	if err := k.ValidateAddress(ctx, msg.Recipient); err != nil {
		return sdkerrors.Wrap(err, "invalid recipient address")
	}

	if _, ok := k.GetMessage(ctx, msg.ID); ok {
		return fmt.Errorf("general message %s already exists", msg.ID)
	}

	if err := k.setMessage(ctx, msg); err != nil {
		return err
	}

	switch msg.Status {
	case exported.Approved:
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageReceived{
			ID:          msg.ID,
			PayloadHash: msg.PayloadHash,
			Sender:      msg.Sender,
			Recipient:   msg.Recipient,
		}))
	case exported.Processing:
		if err := k.setProcessingMessageID(ctx, msg); err != nil {
			return err
		}

		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageProcessing{ID: msg.ID}))
	default:
		return fmt.Errorf("invalid message status %s", msg.Status)
	}

	return nil
}

// SetNewMessage sets the given general message. If the messages is approved, adds the message ID to approved messages store
// Deprecated
func (k Keeper) SetNewMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	sourceChain, ok := k.GetChain(ctx, m.GetSourceChain())
	if !ok {
		return fmt.Errorf("source chain %s is not a registered chain", m.GetSourceChain())
	}

	if err := k.ValidateAddress(ctx, m.Sender); err != nil {
		return err
	}

	destChain, ok := k.GetChain(ctx, m.GetDestinationChain())
	if !ok {
		return fmt.Errorf("destination chain %s is not a registered chain", m.GetDestinationChain())
	}

	if err := k.ValidateAddress(ctx, m.Recipient); err != nil {
		return err
	}

	if m.Asset != nil {
		if err := k.validateAsset(ctx, sourceChain, m.Asset.Denom); err != nil {
			return err
		}

		if err := k.validateAsset(ctx, destChain, m.Asset.Denom); err != nil {
			return err
		}
	}

	if _, found := k.GetMessage(ctx, m.ID); found {
		return fmt.Errorf("general message %s already exists", m.ID)
	}

	if m.Is(exported.Processing) {
		if err := k.setProcessingMessageID(ctx, m); err != nil {
			return err
		}
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageReceived{
		ID:          m.ID,
		PayloadHash: m.PayloadHash,
		Sender:      m.Sender,
		Recipient:   m.Recipient,
	}))

	return k.setMessage(ctx, m)
}

/*
 * Below are the valid message status transitions:
 * Approved -> Processing
 * Processing -> Executed
 * Processing -> Failed
 * Failed -> Processing
 */

// SetMessageProcessing sets the general message as processing
// Deprecated
func (k Keeper) SetMessageProcessing(ctx sdk.Context, id string) error {
	m, found := k.GetMessage(ctx, id)
	if !found {
		return fmt.Errorf("general message %s not found", id)
	}

	if !(m.Is(exported.Approved) || m.Is(exported.Failed)) {
		return fmt.Errorf("general message is not approved or failed")
	}

	m.Status = exported.Processing

	if err := k.setMessage(ctx, m); err != nil {
		return err
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageProcessing{ID: m.ID}))

	return k.setProcessingMessageID(ctx, m)
}

// SetMessageExecuted sets the general message as executed
func (k Keeper) SetMessageExecuted(ctx sdk.Context, id string) error {
	m, found := k.GetMessage(ctx, id)
	if !found {
		return fmt.Errorf("general message %s not found", id)
	}

	if !m.Is(exported.Processing) {
		return fmt.Errorf("general message is not processing")
	}

	k.deleteProcessingMessageID(ctx, m)

	m.Status = exported.Executed

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageExecuted{ID: m.ID}))

	return k.setMessage(ctx, m)
}

// SetMessageFailed sets the general message as failed
func (k Keeper) SetMessageFailed(ctx sdk.Context, id string) error {
	m, found := k.GetMessage(ctx, id)
	if !found {
		return fmt.Errorf("general message %s not found", id)
	}

	if !m.Is(exported.Processing) {
		return fmt.Errorf("general message is not processing")
	}

	k.deleteProcessingMessageID(ctx, m)

	m.Status = exported.Failed

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageFailed{ID: m.ID}))

	return k.setMessage(ctx, m)
}

// GetMessage returns the general message by ID
func (k Keeper) GetMessage(ctx sdk.Context, id string) (m exported.GeneralMessage, found bool) {
	return m, k.getStore(ctx).GetNew(getMessageKey(id), &m)
}

func (k Keeper) setMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getMessageKey(m.ID), &m)
}

func (k Keeper) setProcessingMessageID(ctx sdk.Context, m exported.GeneralMessage) error {
	if !m.Is(exported.Processing) {
		return fmt.Errorf("general message is not processing")
	}

	k.getStore(ctx).SetRawNew(getProcessingMessageKey(m.GetDestinationChain(), m.ID), []byte(m.ID))

	return nil
}

func (k Keeper) deleteProcessingMessageID(ctx sdk.Context, m exported.GeneralMessage) {
	k.getStore(ctx).DeleteNew(getProcessingMessageKey(m.GetDestinationChain(), m.ID))
}

//nolint:unused // TODO: add genesis import/export
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

// GetProcessingMessages returns up to #limit messages that are currently being processed
func (k Keeper) GetProcessingMessages(ctx sdk.Context, chain exported.ChainName, limit int64) []exported.GeneralMessage {
	ids := []string{}

	pageRequest := &query.PageRequest{
		Key:        nil,
		Offset:     0,
		Limit:      uint64(limit),
		CountTotal: false,
		Reverse:    false,
	}
	keyPrefix := append(processingMessagePrefix.Append(key.From(chain)).Bytes(), []byte(key.DefaultDelimiter)...)

	// it's unexpected to get a retrieval/iterator error from IAVL db
	funcs.Must(query.Paginate(prefix.NewStore(k.getStore(ctx).KVStore, keyPrefix), pageRequest, func(key []byte, value []byte) error {
		ids = append(ids, string(value))
		return nil
	}))

	return slices.Map(ids, func(id string) exported.GeneralMessage {
		return funcs.MustOk(k.GetMessage(ctx, id))
	})
}

func (k Keeper) SetNewMessage_(ctx sdk.Context, msg exported.GeneralMessage) error {
	if _, ok := k.GetMessage(ctx, msg.ID); ok {
		return fmt.Errorf("general message %s already exists", msg.ID)
	}

	if !msg.Is(exported.Approved) {
		return fmt.Errorf("new general message has to be approved")
	}

	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageReceived{
		ID:          msg.ID,
		PayloadHash: msg.PayloadHash,
		Sender:      msg.Sender,
		Recipient:   msg.Recipient,
	}))

	return k.setMessage(ctx, msg)
}

func (k Keeper) SetMessageProcessing_(ctx sdk.Context, id string) error {
	msg, ok := k.GetMessage(ctx, id)
	if !ok {
		return fmt.Errorf("general message %s not found", id)
	}

	if !msg.Is(exported.Approved) && !msg.Is(exported.Failed) {
		return fmt.Errorf("general message has to be approved or failed")
	}

	if err := k.validateMessage(ctx, msg); err != nil {
		return err
	}

	msg.Status = exported.Processing
	if err := k.setMessage(ctx, msg); err != nil {
		return err
	}

	funcs.MustNoErr(k.setProcessingMessageID(ctx, msg))
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.MessageProcessing{ID: msg.ID}))

	return nil
}

func (k Keeper) validateMessage(ctx sdk.Context, msg exported.GeneralMessage) error {
	if !msg.Sender.Chain.IsFrom(wasm.ModuleName) {
		if err := k.validateAddressAndAsset(ctx, msg.Sender, msg.Asset); err != nil {
			return err
		}
	}

	if !msg.Recipient.Chain.IsFrom(wasm.ModuleName) {
		if err := k.validateAddressAndAsset(ctx, msg.Recipient, msg.Asset); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) validateAddressAndAsset(ctx sdk.Context, address exported.CrossChainAddress, asset *sdk.Coin) error {
	if _, ok := k.GetChain(ctx, address.Chain.Name); !ok {
		return fmt.Errorf("chain %s is not found", address.Chain.Name)
	}

	if !k.IsChainActivated(ctx, address.Chain) {
		return fmt.Errorf("chain %s is not activated", address.Chain.Name)
	}

	if err := k.ValidateAddress(ctx, address); err != nil {
		return err
	}

	if asset == nil {
		return nil
	}

	return k.validateAsset(ctx, address.Chain, asset.Denom)
}
