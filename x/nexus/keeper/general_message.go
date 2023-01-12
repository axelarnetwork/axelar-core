package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getGeneralMessageKey(chain exported.ChainName, id string) key.Key {
	return generalMessagePrefix.
		Append(key.From(chain)).
		Append(key.FromStr(id))
}

// SetNewGeneralMessage sets the given general message
func (k Keeper) SetNewGeneralMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	sourceChain, ok := k.GetChain(ctx, m.SourceChain)
	if !ok {
		return fmt.Errorf("source chain %s is not a registered chain", m.SourceChain)
	}

	destChain, ok := k.GetChain(ctx, m.DestinationChain)
	if !ok {
		return fmt.Errorf("destination chain %s is not a registered chain", m.DestinationChain)
	}

	if !destChain.IsFrom(axelarnet.ModuleName) {
		return fmt.Errorf("destination chain %s is not a cosmos chain", m.DestinationChain)
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

	if _, found := k.getGeneralMessage(ctx, m.DestinationChain, m.ID); found {
		return fmt.Errorf("general message %s for chain %s already exists", m.ID, m.DestinationChain.String())
	}

	return k.setGeneralMessage(ctx, m)
}

func (k Keeper) getGeneralMessage(ctx sdk.Context, chain exported.ChainName, id string) (m exported.GeneralMessage, found bool) {
	return m, k.getStore(ctx).GetNew(getGeneralMessageKey(chain, id), &m)
}

func (k Keeper) setGeneralMessage(ctx sdk.Context, m exported.GeneralMessage) error {
	return k.getStore(ctx).SetNewValidated(getGeneralMessageKey(m.DestinationChain, m.ID), &m)
}

func (k Keeper) getGeneralMessages(ctx sdk.Context) (generalMessages []exported.GeneralMessage) {
	iter := k.getStore(ctx).IteratorNew(generalMessagePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var generalMessage exported.GeneralMessage
		iter.UnmarshalValue(&generalMessage)

		generalMessages = append(generalMessages, generalMessage)
	}

	return generalMessages
}
