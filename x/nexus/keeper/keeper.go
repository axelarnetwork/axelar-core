package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var (
	senderPrefix     = utils.KeyFromStr("send")
	chainPrefix      = utils.KeyFromStr("chain")
	totalPrefix      = utils.KeyFromStr("total")
	registeredPrefix = utils.KeyFromStr("registered")
	chainStatePrefix = utils.KeyFromStr("chain_state")

	sequenceKey = utils.KeyFromStr("nextID")
	registered  = []byte{0x01}
)

// Keeper represents a nexus keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper returns a new nexus keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the nexus module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)

	for _, chain := range p.Chains {
		// By copying this data to the KV store, we avoid having to iterate across all element
		// in the parameters table when a caller needs to fetch information from it
		k.SetChain(ctx, chain)

		// Native assets can be registered at start up
		k.RegisterAsset(ctx, chain.Name, chain.NativeAsset)
	}
}

// GetParams gets the nexus module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// RegisterAsset indicates that the specified asset is supported by the given chain
func (k Keeper) RegisterAsset(ctx sdk.Context, chainName, denom string) {
	key := registeredPrefix.Append(utils.LowerCaseKey(chainName)).Append(utils.LowerCaseKey(denom))
	k.getStore(ctx).SetRaw(key, registered)
}

// IsAssetRegistered returns true if the specified asset is supported by the given chain
func (k Keeper) IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool {
	key := registeredPrefix.Append(utils.LowerCaseKey(chainName)).Append(utils.LowerCaseKey(denom))
	return k.getStore(ctx).GetRaw(key) != nil
}

// GetChains retrieves the specification for all supported blockchains
func (k Keeper) GetChains(ctx sdk.Context) []exported.Chain {
	var results []exported.Chain

	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), chainPrefix.AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var chain exported.Chain
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &chain)
		results = append(results, chain)
	}

	return results
}

// GetChain retrieves the specification for a supported blockchain
func (k Keeper) GetChain(ctx sdk.Context, chainName string) (exported.Chain, bool) {
	var chain exported.Chain
	key := chainPrefix.Append(utils.LowerCaseKey(chainName))
	ok := k.getStore(ctx).Get(key, &chain)

	return chain, ok
}

// SetChain sets the specification for a supported chain
func (k Keeper) SetChain(ctx sdk.Context, chain exported.Chain) {
	k.getStore(ctx).Set(chainPrefix.Append(utils.LowerCaseKey(chain.Name)), &chain)
}

// LinkAddresses links a sender address to a cross-chain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) {
	k.getStore(ctx).Set(senderPrefix.Append(utils.LowerCaseKey(sender.String())), &recipient)
}

// GetRecipient retrieves the cross chain recipient associated to the specified sender
func (k Keeper) GetRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	var recp exported.CrossChainAddress
	ok := k.getStore(ctx).Get(senderPrefix.Append(utils.LowerCaseKey(sender.String())), &recp)
	return recp, ok
}

// EnqueueForTransfer appoints the amount of tokens to be transfered/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin) error {
	if !sender.Chain.SupportsForeignAssets && sender.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("sender's chain %s does not support foreign assets", sender.Chain.Name)
	}

	if sender.Chain.NativeAsset != asset.Denom && k.getChainTotal(ctx, sender.Chain, asset.Denom).IsLT(asset) {
		return fmt.Errorf("not enough funds available for asset '%s' in chain %s", asset.Denom, sender.Chain.Name)
	}

	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	if !recipient.Chain.SupportsForeignAssets && recipient.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.Name)
	}

	if sender.Chain.NativeAsset != asset.Denom {
		k.subtractFromChainTotal(ctx, sender.Chain, asset)
	}
	k.setPendingTransfer(ctx, recipient, asset)
	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.Amount.String(), recipient.Address, recipient.Chain.Name))

	return nil
}

// ArchivePendingTransfer marks the transfer for the given recipient as concluded and archived
func (k Keeper) ArchivePendingTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer) {
	store := k.getStore(ctx)
	key := utils.LowerCaseKey(transfer.Recipient.Chain.Name).
		Append(utils.LowerCaseKey(strconv.FormatUint(transfer.ID, 10)))
	bz := store.GetRaw(utils.LowerCaseKey(exported.Pending.String()).Append(key))
	if bz == nil {
		return
	}

	// Archive the transfer
	store.Delete(utils.LowerCaseKey(exported.Pending.String()).Append(key))
	store.SetRaw(utils.LowerCaseKey(exported.Archived.String()).Append(key), bz)

	// Update the total nexus for the chain if it is a foreign asset
	var t exported.CrossChainTransfer
	k.cdc.MustUnmarshalLengthPrefixed(bz, &t)
	info, _ := k.GetChain(ctx, t.Recipient.Chain.Name)
	if info.NativeAsset != t.Asset.Denom {
		k.AddToChainTotal(ctx, t.Recipient.Chain, t.Asset)
	}
}

func (k Keeper) getChainTotal(ctx sdk.Context, chain exported.Chain, denom string) sdk.Coin {
	var total sdk.Coin
	ok := k.getStore(ctx).Get(totalPrefix.Append(utils.LowerCaseKey(chain.Name)).Append(utils.LowerCaseKey(denom)), &total)
	if !ok {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}

	return total
}

// AddToChainTotal add balance for an asset for a chain
func (k Keeper) AddToChainTotal(ctx sdk.Context, chain exported.Chain, amount sdk.Coin) {
	total := k.getChainTotal(ctx, chain, amount.Denom)
	total = total.Add(amount)

	k.getStore(ctx).Set(totalPrefix.Append(utils.LowerCaseKey(chain.Name)).Append(utils.LowerCaseKey(amount.Denom)), &total)
}

func (k Keeper) subtractFromChainTotal(ctx sdk.Context, chain exported.Chain, withdrawal sdk.Coin) {
	total := k.getChainTotal(ctx, chain, withdrawal.Denom)
	total = total.Sub(withdrawal)

	k.getStore(ctx).Set(totalPrefix.Append(utils.LowerCaseKey(chain.Name)).Append(utils.LowerCaseKey(withdrawal.Denom)), &total)
}

func (k Keeper) setPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) {
	var next uint64
	store := k.getStore(ctx)
	bz := store.GetRaw(sequenceKey)
	if bz != nil {
		next = binary.LittleEndian.Uint64(bz)
	}

	transfer := exported.CrossChainTransfer{Recipient: recipient, Asset: amount, ID: next}
	key := utils.LowerCaseKey(exported.Pending.String()).
		Append(utils.LowerCaseKey(recipient.Chain.Name)).
		Append(utils.LowerCaseKey(strconv.FormatUint(next, 10)))
	store.Set(key, &transfer)

	next++
	bz = make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, next)
	store.SetRaw(sequenceKey, bz)
}

// GetTransfersForChain returns the current set of transfers with the given state for the given chain
func (k Keeper) GetTransfersForChain(ctx sdk.Context, chain exported.Chain, state exported.TransferState) []exported.CrossChainTransfer {
	transfers := make([]exported.CrossChainTransfer, 0)

	prefix := utils.LowerCaseKey(state.String()).Append(utils.LowerCaseKey(chain.Name))
	iter := k.getStore(ctx).Iterator(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var transfer exported.CrossChainTransfer
		k.cdc.MustUnmarshalLengthPrefixed(bz, &transfer)
		transfers = append(transfers, transfer)
	}

	return transfers
}

func (k Keeper) getChainState(ctx sdk.Context, chain exported.Chain) types.ChainState {
	key := chainStatePrefix.Append(utils.LowerCaseKey(chain.Name))

	var chainState types.ChainState
	if ok := k.getStore(ctx).Get(key, &chainState); ok {
		return chainState
	}

	return types.ChainState{
		Chain:       chain,
		Maintainers: []sdk.ValAddress{},
		Activated:   false,
	}
}

func (k Keeper) setChainState(ctx sdk.Context, chainState types.ChainState) {
	key := chainStatePrefix.Append(utils.LowerCaseKey(chainState.Chain.Name))
	k.getStore(ctx).Set(key, &chainState)
}

// IsChainMaintainer returns true if the given address is one of the given chain's maintainers; false otherwise
func (k Keeper) IsChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) bool {
	return k.getChainState(ctx, chain).HasMaintainer(maintainer)
}

// AddChainMaintainer adds the given address to be one of the given chain's maintainers
func (k Keeper) AddChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) error {
	chainState := k.getChainState(ctx, chain)
	if err := chainState.AddMaintainer(maintainer); err != nil {
		return err
	}

	k.setChainState(ctx, chainState)

	return nil
}

// RemoveChainMaintainer removes the given address from the given chain's maintainers
func (k Keeper) RemoveChainMaintainer(ctx sdk.Context, chain exported.Chain, maintainer sdk.ValAddress) error {
	chainState := k.getChainState(ctx, chain)
	if err := chainState.RemoveMaintainer(maintainer); err != nil {
		return err
	}

	k.setChainState(ctx, chainState)

	return nil
}

// GetChainMaintainers returns the maintainers of the given chain
func (k Keeper) GetChainMaintainers(ctx sdk.Context, chain exported.Chain) []sdk.ValAddress {
	return k.getChainState(ctx, chain).Maintainers
}

// ActivateChain activates the given chain
func (k Keeper) ActivateChain(ctx sdk.Context, chain exported.Chain) {
	chainState := k.getChainState(ctx, chain)
	chainState.Activated = true

	k.setChainState(ctx, chainState)
}

// IsChainActivated returns true if the given chain is activated; false otherwise
func (k Keeper) IsChainActivated(ctx sdk.Context, chain exported.Chain) bool {
	return k.getChainState(ctx, chain).Activated
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
