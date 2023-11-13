package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

func getLatestDepositAddressKey(depositChain exported.ChainName, recipientAddress exported.CrossChainAddress) utils.Key {
	return latestDepositAddressPrefix.
		Append(utils.LowerCaseKey(depositChain.String())).
		Append(utils.LowerCaseKey(recipientAddress.Chain.Name.String())).
		Append(utils.LowerCaseKey(recipientAddress.Address))
}

func (k Keeper) setLatestDepositAddress(ctx sdk.Context, recipientAddress, depositAddress exported.CrossChainAddress) {
	k.getStore(ctx).Set(getLatestDepositAddressKey(depositAddress.Chain.Name, recipientAddress), &depositAddress)
}

func (k Keeper) getLatestDepositAddress(ctx sdk.Context, depositChain exported.ChainName, recipientAddress exported.CrossChainAddress) (depositAddress exported.CrossChainAddress, ok bool) {
	return depositAddress, k.getStore(ctx).Get(getLatestDepositAddressKey(depositChain, recipientAddress), &depositAddress)
}

func getLinkedAddressesKey(depositAddress exported.CrossChainAddress) utils.Key {
	return linkedAddressesPrefix.
		Append(utils.LowerCaseKey(depositAddress.Chain.Name.String())).
		Append(utils.LowerCaseKey(depositAddress.Address))
}

func (k Keeper) setLinkedAddresses(ctx sdk.Context, linkedAddresses types.LinkedAddresses) {
	k.getStore(ctx).Set(getLinkedAddressesKey(linkedAddresses.DepositAddress), &linkedAddresses)
}

func (k Keeper) getLinkedAddresses(ctx sdk.Context, depositAddress exported.CrossChainAddress) (linkedAddresses types.LinkedAddresses, ok bool) {
	return linkedAddresses, k.getStore(ctx).Get(getLinkedAddressesKey(depositAddress), &linkedAddresses)
}

func (k Keeper) getAllLinkedAddresses(ctx sdk.Context) (results []types.LinkedAddresses) {
	iter := k.getStore(ctx).Iterator(linkedAddressesPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var linkedAddresses types.LinkedAddresses
		iter.UnmarshalValue(&linkedAddresses)

		results = append(results, linkedAddresses)
	}

	return results
}

// LinkAddresses links a sender address to a cross-chain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, depositAddress exported.CrossChainAddress, recipientAddress exported.CrossChainAddress) error {
	if err := k.ValidateAddress(ctx, depositAddress); err != nil {
		return err
	}

	if err := k.ValidateAddress(ctx, recipientAddress); err != nil {
		return err
	}

	if !k.IsChainActivated(ctx, depositAddress.Chain) {
		return fmt.Errorf("sender chain '%s' is not activated", depositAddress.Chain.Name)
	}

	if !k.IsChainActivated(ctx, recipientAddress.Chain) {
		return fmt.Errorf("recipient chain '%s' is not activated", recipientAddress.Chain.Name)
	}

	linkedAddresses := types.NewLinkedAddresses(depositAddress, recipientAddress)

	k.setLinkedAddresses(ctx, linkedAddresses)
	k.setLatestDepositAddress(ctx, recipientAddress, depositAddress)

	return nil
}

// GetRecipient retrieves the cross chain recipient associated to the specified sender
func (k Keeper) GetRecipient(ctx sdk.Context, depositAddress exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	if linkedAddresses, ok := k.getLinkedAddresses(ctx, depositAddress); ok {
		return linkedAddresses.RecipientAddress, true
	}

	return exported.CrossChainAddress{}, false
}

// ValidateAddress validates the given cross chain address
func (k Keeper) ValidateAddress(ctx sdk.Context, address exported.CrossChainAddress) error {
	validate, err := k.getAddressValidator(address.Chain.Module)
	if err != nil {
		return err
	}

	if err := validate(ctx, address); err != nil {
		return err
	}

	return nil
}
