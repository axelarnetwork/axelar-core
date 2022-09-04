package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

func getLatestDepositAddressKey(depositChain exported.ChainName, recipientAddress exported.CrossChainAddress) key.Key {
	return latestDepositAddressPrefix.
		Append(key.From(depositChain)).
		Append(key.From(recipientAddress.Chain.Name)).
		Append(key.FromStr(recipientAddress.Address))
}

func (k Keeper) setLatestDepositAddress(ctx sdk.Context, recipientAddress, depositAddress exported.CrossChainAddress) {
	k.getStore(ctx).SetNew(getLatestDepositAddressKey(depositAddress.Chain.Name, recipientAddress), &depositAddress)
}

func (k Keeper) getLatestDepositAddress(ctx sdk.Context, depositChain exported.ChainName, recipientAddress exported.CrossChainAddress) (depositAddress exported.CrossChainAddress, ok bool) {
	return depositAddress, k.getStore(ctx).GetNew(getLatestDepositAddressKey(depositChain, recipientAddress), &depositAddress)
}

func getLinkedAddressesKey(depositAddress exported.CrossChainAddress) key.Key {
	return linkedAddressesPrefix.
		Append(key.From(depositAddress.Chain.Name)).
		Append(key.FromStr(depositAddress.Address))
}

func (k Keeper) setLinkedAddresses(ctx sdk.Context, linkedAddresses types.LinkedAddresses) {
	k.getStore(ctx).SetNew(getLinkedAddressesKey(linkedAddresses.DepositAddress), &linkedAddresses)
}

func (k Keeper) getLinkedAddresses(ctx sdk.Context, depositAddress exported.CrossChainAddress) (linkedAddresses types.LinkedAddresses, ok bool) {
	return linkedAddresses, k.getStore(ctx).GetNew(getLinkedAddressesKey(depositAddress), &linkedAddresses)
}

func (k Keeper) getAllLinkedAddresses(ctx sdk.Context) []types.LinkedAddresses {
	return utils.GetValues[types.LinkedAddresses](k.getStore(ctx), linkedAddressesPrefix)
}

// LinkAddresses links a sender address to a cross-chain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, depositAddress exported.CrossChainAddress, recipientAddress exported.CrossChainAddress) error {
	if validator := k.GetRouter().GetAddressValidator(depositAddress.Chain.Module); validator == nil {
		return fmt.Errorf("unknown module for sender's chain %s", depositAddress.Chain.String())
	} else if err := validator(ctx, depositAddress); err != nil {
		return err
	}

	if validator := k.GetRouter().GetAddressValidator(recipientAddress.Chain.Module); validator == nil {
		return fmt.Errorf("unknown module for recipient's chain %s", recipientAddress.Chain.String())
	} else if err := validator(ctx, recipientAddress); err != nil {
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
