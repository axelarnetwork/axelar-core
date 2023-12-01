package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	"github.com/axelarnetwork/utils/funcs"
)

// UndelegateDecorator checks if the unbonding validator holds any multiSig share of active crypto keys
type UndelegateDecorator struct {
	multiSig    types.MultiSig
	nexus       types.Nexus
	snapshotter types.Snapshotter
}

// NewUndelegateDecorator constructor for UndelegateDecorator
func NewUndelegateDecorator(multiSig types.MultiSig, nexus types.Nexus, snapshotter types.Snapshotter) UndelegateDecorator {
	return UndelegateDecorator{
		multiSig,
		nexus,
		snapshotter,
	}
}

// AnteHandle fails the transaction if it finds any validator holding multiSig share of active keys is trying to unbond
func (d UndelegateDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgUndelegate:
			valAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			// only restrict a validator from unbonding it's self-delegation
			if !delegatorAddress.Equals(valAddress) {
				continue
			}

			chains := d.nexus.GetChains(ctx)

			for _, chain := range chains {
				if !d.nexus.IsChainActivated(ctx, chain) {
					continue
				}

				nextKeyID, idFound := d.multiSig.GetNextKeyID(ctx, chain.Name)
				key, keyFound := d.multiSig.GetKey(ctx, nextKeyID)
				if idFound && keyFound && !key.GetWeight(valAddress).IsZero() {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding multiSig share of %s's next key %s", valAddress, chain.Name, nextKeyID)
				}

				activeKeyIDs := d.multiSig.GetActiveKeyIDs(ctx, chain.Name)
				for _, activeKeyID := range activeKeyIDs {
					key := funcs.MustOk(d.multiSig.GetKey(ctx, activeKeyID))
					if !key.GetWeight(valAddress).IsZero() {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding multiSig share of %s's active key %s", valAddress, chain.Name, activeKeyID)
					}
				}
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}
