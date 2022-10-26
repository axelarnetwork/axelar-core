package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var (
	cosmosChainPrefix = key.FromStr("cosmos_chain")
	feeCollector      = key.FromStr("fee_collector")

	transferPrefix       = key.FromStr("ibc_transfer")
	ibcTransferQueueName = "route_transfer_queue"
	// nonceKey is deprecated in v0.23
	// nonceKey = key.FromUInt[uint64](1) <-- keeping this here so the number doesn't get reused

	// failedTransferPrefix is deprecated in v0.23
	failedTransferPrefix = key.FromUInt[uint64](2)
	seqIDMappingPrefix   = key.FromUInt[uint64](3)
)

// Keeper provides access to all state changes regarding the Axelarnet module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace

	channelK types.ChannelKeeper
}

// NewKeeper returns a new axelarnet keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace, channelK types.ChannelKeeper) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable()), channelK: channelK}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) getParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

func (k Keeper) setParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// GetRouteTimeoutWindow returns the timeout window for IBC transfers routed by axelarnet
func (k Keeper) GetRouteTimeoutWindow(ctx sdk.Context) uint64 {
	var result uint64
	k.params.Get(ctx, types.KeyRouteTimeoutWindow, &result)

	return result
}

// GetTransferLimit returns the transfer limit for transfers processed by axelarnet
func (k Keeper) GetTransferLimit(ctx sdk.Context) uint64 {
	var result uint64
	k.params.Get(ctx, types.KeyTransferLimit, &result)

	return result
}

// GetEndBlockerLimit returns the transfer limit for IBC transfers routed in the end blocker by axelarnet
func (k Keeper) GetEndBlockerLimit(ctx sdk.Context) uint64 {
	var result uint64
	k.params.Get(ctx, types.KeyEndBlockerLimit, &result)

	return result
}

// GetIBCPath retrieves the IBC path associated to the specified chain
func (k Keeper) GetIBCPath(ctx sdk.Context, chain nexus.ChainName) (string, bool) {
	cosmosChain, ok := k.getCosmosChain(ctx, chain)
	if !ok || cosmosChain.IBCPath == "" {
		return "", false
	}

	return cosmosChain.IBCPath, true
}

// IsCosmosChain returns true if the given chain name is for a cosmos chain
func (k Keeper) IsCosmosChain(ctx sdk.Context, chain nexus.ChainName) bool {
	_, ok := k.getCosmosChain(ctx, chain)
	return ok
}

// GetCosmosChainByName gets the address prefix of the given cosmos chain
func (k Keeper) GetCosmosChainByName(ctx sdk.Context, chain nexus.ChainName) (types.CosmosChain, bool) {
	chainKey := cosmosChainPrefix.Append(key.From(chain))
	var value types.CosmosChain
	ok := k.getStore(ctx).GetNew(chainKey, &value)
	if !ok {
		return types.CosmosChain{}, false
	}

	return value, true
}

// GetCosmosChains retrieves all registered cosmos chain names
func (k Keeper) GetCosmosChains(ctx sdk.Context) []nexus.ChainName {
	return slices.Map(k.getCosmosChains(ctx), func(c types.CosmosChain) nexus.ChainName { return c.Name })
}

func (k Keeper) getCosmosChains(ctx sdk.Context) (cosmosChains []types.CosmosChain) {
	iter := k.getStore(ctx).IteratorNew(cosmosChainPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var cosmosChain types.CosmosChain
		iter.UnmarshalValue(&cosmosChain)

		cosmosChains = append(cosmosChains, cosmosChain)
	}

	return cosmosChains
}

func (k Keeper) getCosmosChain(ctx sdk.Context, chain nexus.ChainName) (cosmosChain types.CosmosChain, ok bool) {
	return cosmosChain, k.getStore(ctx).GetNew(cosmosChainPrefix.Append(key.From(chain)), &cosmosChain)
}

// SetCosmosChain sets the address prefix for the given cosmos chain
func (k Keeper) SetCosmosChain(ctx sdk.Context, chain types.CosmosChain) error {
	// register a cosmos chain to axelarnet
	return k.getStore(ctx).SetNewValidated(cosmosChainPrefix.Append(key.From(chain.Name)), &chain)
}

// SetFeeCollector sets axelarnet fee collector
func (k Keeper) SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) error {
	if err := sdk.VerifyAddressFormat(address); err != nil {
		return err
	}

	k.getStore(ctx).SetRawNew(feeCollector, address)
	return nil
}

// GetFeeCollector gets axelarnet fee collector
func (k Keeper) GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool) {
	bz := k.getStore(ctx).GetRawNew(feeCollector)
	if bz == nil {
		return sdk.AccAddress{}, false
	}

	return bz, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

// GetIBCTransferQueue returns the queue of IBC transfers
func (k Keeper) GetIBCTransferQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewGeneralKVQueue(
		ibcTransferQueueName,
		k.getStore(ctx),
		k.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			transfer := value.(*types.IBCTransfer)
			return utils.KeyFromBz(transfer.ID.Bytes())
		},
	)
}

func getTransferKey(id nexus.TransferID) key.Key {
	return transferPrefix.Append(key.From(id))
}

// EnqueueIBCTransfer stores the pending ibc transfer in the queue
func (k Keeper) EnqueueIBCTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	transferKey := getTransferKey(transfer.ID)
	if k.getStore(ctx).HasNew(transferKey) {
		return fmt.Errorf("transfer %s already exists", transfer.ID.String())
	}

	k.GetIBCTransferQueue(ctx).Enqueue(utils.KeyFromBz(transferKey.Bytes()), &transfer)
	return nil
}

// validateIBCTransferQueueState checks if the keys of the given map have the correct format to be imported as ibc transfer queue state.
func (k Keeper) validateIBCTransferQueueState(state utils.QueueState, queueName ...string) error {
	if err := state.ValidateBasic(queueName...); err != nil {
		return err
	}

	for _, item := range state.Items {
		var transfer types.IBCTransfer
		if err := k.cdc.UnmarshalLengthPrefixed(item.Value, &transfer); err != nil {
			return err
		}

		if err := transfer.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func getFailedTransferKey(id nexus.TransferID) key.Key {
	return failedTransferPrefix.Append(key.FromBz(id.Bytes()))
}

// GetFailedTransfer returns the failed transfer for the given transfer ID
func (k Keeper) GetFailedTransfer(ctx sdk.Context, id nexus.TransferID) (transfer types.IBCTransfer, ok bool) {
	return transfer, k.getStore(ctx).GetNew(getFailedTransferKey(id), &transfer)
}

// GetTransfer returns the ibc transfer for the given transfer ID
func (k Keeper) GetTransfer(ctx sdk.Context, id nexus.TransferID) (transfer types.IBCTransfer, ok bool) {
	k.getStore(ctx).GetNew(getTransferKey(id), &transfer)
	return transfer, transfer.Status != types.TransferNonExistent
}

func (k Keeper) setTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	return k.getStore(ctx).SetNewValidated(getTransferKey(transfer.ID), &transfer)
}

func (k Keeper) getFailedTransfers(ctx sdk.Context) (failedTransfers []types.IBCTransfer) {
	iter := k.getStore(ctx).IteratorNew(failedTransferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var t types.IBCTransfer
		iter.UnmarshalValue(&t)

		failedTransfers = append(failedTransfers, t)
	}

	return failedTransfers
}

func (k Keeper) setTransferStatus(ctx sdk.Context, transferID nexus.TransferID, status types.IBCTransfer_Status) error {
	t, ok := k.GetTransfer(ctx, transferID)
	if !ok {
		return fmt.Errorf("transfer %s not found", transferID)
	}

	err := t.SetStatus(status)
	if err != nil {
		return err
	}

	return k.setTransfer(ctx, t)
}

// SetTransferCompleted sets the transfer as completed
func (k Keeper) SetTransferCompleted(ctx sdk.Context, transferID nexus.TransferID) error {
	return k.setTransferStatus(ctx, transferID, types.TransferCompleted)
}

// SetTransferFailed sets the transfer as failed
func (k Keeper) SetTransferFailed(ctx sdk.Context, transferID nexus.TransferID) error {
	return k.setTransferStatus(ctx, transferID, types.TransferFailed)
}

// SetTransferPending sets the transfer as pending
func (k Keeper) SetTransferPending(ctx sdk.Context, transferID nexus.TransferID) error {
	return k.setTransferStatus(ctx, transferID, types.TransferPending)
}

func getSeqIDMappingKey(portID, channelID string, seq uint64) key.Key {
	return seqIDMappingPrefix.
		Append(key.FromStr(portID)).
		Append(key.FromStr(channelID)).
		Append(key.FromUInt(seq))
}

// SetSeqIDMapping sets transfer ID by port, channel and packet seq
func (k Keeper) SetSeqIDMapping(ctx sdk.Context, t types.IBCTransfer) error {
	nextSeq, ok := k.channelK.GetNextSequenceSend(ctx, t.PortID, t.ChannelID)
	if !ok {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", t.PortID, t.ChannelID,
		)
	}
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(
			getSeqIDMappingKey(t.PortID, t.ChannelID, nextSeq),
			utils.NoValidation(&gogoprototypes.UInt64Value{Value: uint64(t.ID)}),
		),
	)

	return nil
}

// GetSeqIDMapping gets transfer ID by port, channel and packet seq
func (k Keeper) GetSeqIDMapping(ctx sdk.Context, portID, channelID string, seq uint64) (nexus.TransferID, bool) {
	var val gogoprototypes.UInt64Value
	return nexus.TransferID(val.Value), k.getStore(ctx).GetNew(getSeqIDMappingKey(portID, channelID, seq), &val)
}

// DeleteSeqIDMapping deletes (port, channel, packet seq) -> transfer ID mapping
func (k Keeper) DeleteSeqIDMapping(ctx sdk.Context, portID, channelID string, seq uint64) {
	k.getStore(ctx).DeleteRaw(getSeqIDMappingKey(portID, channelID, seq).Bytes())
}

func (k Keeper) getIBCTransfers(ctx sdk.Context) (transfers []types.IBCTransfer) {
	iter := k.getStore(ctx).IteratorNew(transferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var t types.IBCTransfer
		iter.UnmarshalValue(&t)

		transfers = append(transfers, t)
	}

	return transfers
}

func (k Keeper) getSeqIDMappings(ctx sdk.Context) map[string]uint64 {
	mapping := make(map[string]uint64)

	iter := k.getStore(ctx).IteratorNew(seqIDMappingPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var val gogoprototypes.UInt64Value
		iter.UnmarshalValue(&val)

		mapping[string(iter.Key())] = val.Value
	}

	return mapping
}
