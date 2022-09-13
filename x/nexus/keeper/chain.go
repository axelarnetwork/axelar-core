package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func (k Keeper) getChainStates(ctx sdk.Context) (chainStates []types.ChainState) {
	iter := k.getStore(ctx).Iterator(chainStatePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var chainState types.ChainState
		iter.UnmarshalValue(&chainState)

		chainStates = append(chainStates, chainState)
	}

	return chainStates
}

func (k Keeper) getChainState(ctx sdk.Context, chain exported.Chain) (chainState types.ChainState, ok bool) {
	return chainState, k.getStore(ctx).Get(chainStatePrefix.Append(utils.LowerCaseKey(chain.Name.String())), &chainState)
}

// RegisterAsset indicates that the specified asset is supported by the given chain
func (k Keeper) RegisterAsset(ctx sdk.Context, chain exported.Chain, asset exported.Asset) error {
	chainState, _ := k.getChainState(ctx, chain)
	chainState.Chain = chain

	if asset.IsNativeAsset {
		if c, ok := k.GetChainByNativeAsset(ctx, asset.Denom); ok {
			return fmt.Errorf("native asset %s already set for chain %s", asset.Denom, c.Name)
		}
		k.setChainByNativeAsset(ctx, asset.Denom, chain)
	}

	if err := chainState.AddAsset(asset); err != nil {
		return err
	}

	k.setChainState(ctx, chainState)

	return nil
}

// IsAssetRegistered returns true if the specified asset is supported by the given chain
func (k Keeper) IsAssetRegistered(ctx sdk.Context, chain exported.Chain, denom string) bool {
	chainState, ok := k.getChainState(ctx, chain)
	if !ok {
		return false
	}

	return chainState.HasAsset(denom)
}

func (k Keeper) getFeeInfos(ctx sdk.Context) (feeInfos []exported.FeeInfo) {
	iter := k.getStore(ctx).Iterator(assetFeePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var feeInfo exported.FeeInfo
		iter.UnmarshalValue(&feeInfo)

		feeInfos = append(feeInfos, feeInfo)
	}

	return feeInfos
}

func (k Keeper) setFeeInfo(ctx sdk.Context, chain exported.Chain, asset string, feeInfo exported.FeeInfo) {
	k.getStore(ctx).Set(assetFeePrefix.Append(utils.LowerCaseKey(chain.Name.String())).Append(utils.KeyFromStr(asset)), &feeInfo)
}

// GetFeeInfo retrieves the fee info for an asset on a chain, and returns zero fees if it doesn't exist
func (k Keeper) GetFeeInfo(ctx sdk.Context, chain exported.Chain, asset string) (feeInfo exported.FeeInfo, found bool) {
	found = k.getStore(ctx).Get(assetFeePrefix.Append(utils.LowerCaseKey(chain.Name.String())).Append(utils.KeyFromStr(asset)), &feeInfo)
	if !found {
		feeInfo = exported.ZeroFeeInfo(chain.Name, asset)
	}
	return feeInfo, found
}

// RegisterFee registers the fee info for an asset on a chain
func (k Keeper) RegisterFee(ctx sdk.Context, chain exported.Chain, feeInfo exported.FeeInfo) error {
	asset := feeInfo.Asset

	if !k.IsAssetRegistered(ctx, chain, asset) {
		return fmt.Errorf("%s is not a registered asset for chain %s", asset, chain.Name)
	}

	k.setFeeInfo(ctx, chain, asset, feeInfo)

	return nil
}

// ActivateChain activates the given chain
func (k Keeper) ActivateChain(ctx sdk.Context, chain exported.Chain) {
	chainState, _ := k.getChainState(ctx, chain)
	chainState.Chain = chain
	chainState.Activated = true

	k.setChainState(ctx, chainState)
}

// DeactivateChain deactivates the given chain
func (k Keeper) DeactivateChain(ctx sdk.Context, chain exported.Chain) {
	chainState, _ := k.getChainState(ctx, chain)
	chainState.Chain = chain
	chainState.Activated = false

	k.setChainState(ctx, chainState)
}

// IsChainActivated returns true if the given chain is activated; false otherwise
func (k Keeper) IsChainActivated(ctx sdk.Context, chain exported.Chain) bool {
	chainState, ok := k.getChainState(ctx, chain)
	if !ok {
		return false
	}

	return chainState.Activated
}

// GetChainMaintainerState returns the maintainer state of the given chain and address
func (k Keeper) GetChainMaintainerState(ctx sdk.Context, chain exported.Chain, address sdk.ValAddress) (exported.MaintainerState, bool) {
	ms, ok := k.getChainMaintainerState(ctx, chain.Name, address)
	if !ok {
		return nil, false
	}

	return &ms, true
}

// GetChainMaintainerStates returns the maintainer states of the given chain
func (k Keeper) GetChainMaintainerStates(ctx sdk.Context, chain exported.Chain) []exported.MaintainerState {
	return slices.Map(k.getChainMaintainerStates(ctx, chain.Name), func(ms types.MaintainerState) exported.MaintainerState {
		return &ms
	})
}

// SetChainMaintainerState sets the given chain's maintainer state
func (k Keeper) SetChainMaintainerState(ctx sdk.Context, maintainerState exported.MaintainerState) error {
	ms := maintainerState.(*types.MaintainerState)

	if !k.hasChainMaintainerState(ctx, ms.Chain, ms.Address) {
		return fmt.Errorf("%s is not a chain maintainer of chain %s", ms.Address.String(), ms.Chain.String())
	}

	k.setChainMaintainerState(ctx, ms)

	return nil
}

// GetChainMaintainers returns the maintainers of the given chain
func (k Keeper) GetChainMaintainers(ctx sdk.Context, chain exported.Chain) []sdk.ValAddress {
	return slices.Map(k.getChainMaintainerStates(ctx, chain.Name), types.MaintainerState.GetAddress)
}

// IsChainMaintainer returns true if the given address is one of the given chain's maintainers; false otherwise
func (k Keeper) IsChainMaintainer(ctx sdk.Context, chain exported.Chain, address sdk.ValAddress) bool {
	return k.hasChainMaintainerState(ctx, chain.Name, address)
}

// AddChainMaintainer adds the given address to be one of the given chain's maintainers
func (k Keeper) AddChainMaintainer(ctx sdk.Context, chain exported.Chain, address sdk.ValAddress) error {
	if k.hasChainMaintainerState(ctx, chain.Name, address) {
		return fmt.Errorf("%s is already a chain maintainer of chain %s", address.String(), chain.Name.String())
	}

	k.setChainMaintainerState(ctx, types.NewMaintainerState(chain.Name, address))

	return nil
}

// RemoveChainMaintainer removes the given address from the given chain's maintainers
func (k Keeper) RemoveChainMaintainer(ctx sdk.Context, chain exported.Chain, address sdk.ValAddress) error {
	if !k.hasChainMaintainerState(ctx, chain.Name, address) {
		return fmt.Errorf("%s is not a chain maintainer of chain %s", address.String(), chain.Name.String())
	}

	k.deleteChainMaintainerState(ctx, chain.Name, address)

	return nil
}

// GetChains retrieves the specification for all supported blockchains
func (k Keeper) GetChains(ctx sdk.Context) (chains []exported.Chain) {
	iter := k.getStore(ctx).Iterator(chainPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var chain exported.Chain
		iter.UnmarshalValue(&chain)

		chains = append(chains, chain)
	}

	return chains
}

// GetChain retrieves the specification for a supported blockchain
func (k Keeper) GetChain(ctx sdk.Context, chainName exported.ChainName) (chain exported.Chain, ok bool) {
	return chain, k.getStore(ctx).Get(chainPrefix.Append(utils.LowerCaseKey(chainName.String())), &chain)
}

// SetChain sets the specification for a supported chain
func (k Keeper) SetChain(ctx sdk.Context, chain exported.Chain) {
	k.getStore(ctx).Set(chainPrefix.Append(utils.LowerCaseKey(chain.Name.String())), &chain)
}

func (k Keeper) setChainByNativeAsset(ctx sdk.Context, asset string, chain exported.Chain) {
	k.getStore(ctx).Set(chainByNativeAssetPrefix.Append(utils.LowerCaseKey(asset)), &chain)
}

// GetChainByNativeAsset gets a chain by the native asset
func (k Keeper) GetChainByNativeAsset(ctx sdk.Context, asset string) (chain exported.Chain, ok bool) {
	return chain, k.getStore(ctx).Get(chainByNativeAssetPrefix.Append(utils.LowerCaseKey(asset)), &chain)
}

func (k Keeper) setChainState(ctx sdk.Context, chainState types.ChainState) {
	k.getStore(ctx).Set(chainStatePrefix.Append(utils.LowerCaseKey(chainState.ChainName().String())), &chainState)
}

func (k Keeper) getChainMaintainerStates(ctx sdk.Context, chain exported.ChainName) []types.MaintainerState {
	var results []types.MaintainerState

	iter := k.getStore(ctx).IteratorNew(chainMaintainerStatePrefix.Append(key.From(chain)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var maintainerState types.MaintainerState
		iter.UnmarshalValue(&maintainerState)

		results = append(results, maintainerState)
	}

	return results
}

func (k Keeper) getChainMaintainerState(ctx sdk.Context, chain exported.ChainName, address sdk.ValAddress) (ms types.MaintainerState, ok bool) {
	return ms, k.getStore(ctx).GetNew(chainMaintainerStatePrefix.Append(key.From(chain)).Append(key.FromBz(address.Bytes())), &ms)
}

func (k Keeper) setChainMaintainerState(ctx sdk.Context, maintainerState *types.MaintainerState) {
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(
			chainMaintainerStatePrefix.Append(key.From(maintainerState.Chain)).Append(key.FromBz(maintainerState.Address.Bytes())),
			maintainerState))
}

func (k Keeper) deleteChainMaintainerState(ctx sdk.Context, chain exported.ChainName, address sdk.ValAddress) {
	k.getStore(ctx).DeleteNew(chainMaintainerStatePrefix.Append(key.From(chain)).Append(key.FromBz(address.Bytes())))
}

func (k Keeper) hasChainMaintainerState(ctx sdk.Context, chain exported.ChainName, address sdk.ValAddress) bool {
	return k.getStore(ctx).HasNew(chainMaintainerStatePrefix.Append(key.From(chain)).Append(key.FromBz(address.Bytes())))
}
