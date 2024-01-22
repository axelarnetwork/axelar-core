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

const routeMessageQueue = "route_message_queue"

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

// SetNewMessage sets the given general messsage as approved
func (k Keeper) SetNewMessage(ctx sdk.Context, msg exported.GeneralMessage) error {
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

// setMessageProcessing sets the given general message as processing and perform
// validations on the message
func (k Keeper) setMessageProcessing(ctx sdk.Context, id string) error {
	msg, ok := k.GetMessage(ctx, id)
	if !ok {
		return fmt.Errorf("general message %s not found", id)
	}

	if !(msg.Is(exported.Approved) || msg.Is(exported.Failed)) {
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
	// only validate sender and asset if it's not from wasm.
	// the nexus module doesn't know how to validate wasm chains and addresses.
	if !msg.Sender.Chain.IsFrom(wasm.ModuleName) {
		if err := k.validateAddressAndAsset(ctx, msg.Sender, msg.Asset); err != nil {
			return err
		}
	}

	// only validate recipient and asset if it's not to wasm.
	// the nexus module doesn't know how to validate wasm chains and addresses.
	if !msg.Recipient.Chain.IsFrom(wasm.ModuleName) {
		if err := k.validateAddressAndAsset(ctx, msg.Recipient, msg.Asset); err != nil {
			return err
		}
	}

	// asset is not supported for wasm messages
	if (msg.Sender.Chain.IsFrom(wasm.ModuleName) || msg.Recipient.Chain.IsFrom(wasm.ModuleName)) && msg.Asset != nil {
		return fmt.Errorf("asset transfer is not supported for wasm messages")
	}

	return nil
}

// validateAddressAndAsset validates 1) chain existence, 2) chain activation, 3) address, 4) asset
func (k Keeper) validateAddressAndAsset(ctx sdk.Context, address exported.CrossChainAddress, asset *sdk.Coin) error {
	if _, ok := k.GetChain(ctx, address.Chain.Name); !ok {
		return fmt.Errorf("chain %s is not registered", address.Chain.Name)
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

func (k Keeper) getRouteMessageQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewBlockHeightKVQueue(routeMessageQueue, k.getStore(ctx), ctx.BlockHeight(), k.Logger(ctx))
}

// EnqueueRouteMessage enqueues the given general message to be routed
func (k Keeper) EnqueueRouteMessage(ctx sdk.Context, id string) error {
	msg, ok := k.GetMessage(ctx, id)
	if !ok {
		return fmt.Errorf("general message %s not found", id)
	}

	if !(msg.Is(exported.Approved) || msg.Is(exported.Failed)) {
		return fmt.Errorf("general message has to be approved or failed")
	}

	k.getRouteMessageQueue(ctx).Enqueue(utils.KeyFromBz(getMessageKey(id).Bytes()), &msg)

	return nil
}

// DequeueRouteMessage dequeues the next general message to be routed
func (k Keeper) DequeueRouteMessage(ctx sdk.Context) (msg exported.GeneralMessage, ok bool) {
	return msg, k.getRouteMessageQueue(ctx).Dequeue(&msg)
}

// RouteMessage routes the given general message to the corresponding module and
// set the message status to processing
func (k Keeper) RouteMessage(ctx sdk.Context, id string, routingCtx ...exported.RoutingContext) error {
	err := k.setMessageProcessing(ctx, id)
	if err != nil {
		return err
	}

	k.Logger(ctx).Debug("set general message status to processing", "messageID", id)

	if len(routingCtx) == 0 {
		routingCtx = []exported.RoutingContext{{}}
	}

	msg := funcs.MustOk(k.GetMessage(ctx, id))
	if err := k.getMessageRouter().Route(ctx, routingCtx[0], msg); err != nil {
		return sdkerrors.Wrapf(err, "failed to route message %s to the %s module", id, msg.Recipient.Chain.Module)
	}

	return nil
}
