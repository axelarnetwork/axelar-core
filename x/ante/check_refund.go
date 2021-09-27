package ante

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// CheckRefundFeeDecorator record potential refund for tss and vote txs
type CheckRefundFeeDecorator struct {
	ak          antetypes.AccountKeeper
	staking     types.Staking
	axelarnet   types.Axelarnet
	snapshotter types.Snapshotter
	registry    cdctypes.InterfaceRegistry
}

// NewCheckRefundFeeDecorator constructor for CheckRefundFeeDecorator
func NewCheckRefundFeeDecorator(ak antetypes.AccountKeeper, staking types.Staking, snapshotter types.Snapshotter, axelarnet types.Axelarnet, registry cdctypes.InterfaceRegistry, ) CheckRefundFeeDecorator {
	return CheckRefundFeeDecorator{
		ak,
		staking,
		axelarnet,
		snapshotter,
		registry,
	}
}

// AnteHandle record qualified refund for the tss and vote transactions
func (d CheckRefundFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	if d.qualifyForRefund(ctx, msgs) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
		}
		fee := feeTx.GetFee()

		innerMsg := msgs[0].(*axelarnetTypes.RefundMessageRequest).GetInnerMessage()
		err := d.axelarnet.SetPotentialRefund(ctx, axelarnetTypes.GetMsgKey(innerMsg), fee[0])
		if err != nil {
			return ctx, err
		}

	}

	return next(ctx, tx, simulate)
}

func (d CheckRefundFeeDecorator) qualifyForRefund(ctx sdk.Context, msgs []sdk.Msg) bool {
	if len(msgs) != 1 {
		return false
	}

	switch msg := msgs[0].(type) {
	case *axelarnetTypes.RefundMessageRequest:
		if msgRegistered(d.registry, msg.InnerMessage.TypeUrl) {
			// Validator must be bounded
			validatorAddr := getValidator(ctx, d.snapshotter, msg)
			if validatorAddr == nil {
				return false
			}
			validator := d.staking.Validator(ctx, validatorAddr)
			if !validator.IsBonded() {
				return false
			}
		}
	default:
		return false
	}

	return true
}

// getValidator returns the validator address associated to the proxy address
func getValidator(ctx sdk.Context, snapshotter types.Snapshotter, msg sdk.Msg) sdk.ValAddress {
	sender := msg.GetSigners()[0]
	validator := snapshotter.GetOperator(ctx, sender)
	return validator
}

func msgRegistered(r cdctypes.InterfaceRegistry, targetURL string) bool {
	for _, url := range r.ListImplementations("axelarnet.v1beta1.Refundable") {
		if targetURL == url {
			return true
		}
	}
	return false
}
