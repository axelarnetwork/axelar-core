package ante

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// CheckRefundFeeDecorator record potential refund for tss and vote txs
type CheckRefundFeeDecorator struct {
	registry    cdctypes.InterfaceRegistry
	ak          antetypes.AccountKeeper
	staking     types.Staking
	axelarnet   types.Axelarnet
	snapshotter types.Snapshotter
}

// NewCheckRefundFeeDecorator constructor for CheckRefundFeeDecorator
func NewCheckRefundFeeDecorator(registry cdctypes.InterfaceRegistry, ak antetypes.AccountKeeper, staking types.Staking, snapshotter types.Snapshotter, axelarnet types.Axelarnet) CheckRefundFeeDecorator {
	return CheckRefundFeeDecorator{
		registry,
		ak,
		staking,
		axelarnet,
		snapshotter,
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
		if len(fee) > 0 {
			req := msgs[0].(*axelarnetTypes.RefundMsgRequest)
			err := d.axelarnet.SetPendingRefund(ctx, *req, fee[0])
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}
		}

	}

	return next(ctx, tx, simulate)
}

func (d CheckRefundFeeDecorator) qualifyForRefund(ctx sdk.Context, msgs []sdk.Msg) bool {
	if len(msgs) != 1 {
		return false
	}

	switch msg := msgs[0].(type) {
	case *axelarnetTypes.RefundMsgRequest:
		if msgRegistered(d.registry, msg.InnerMessage.GetTypeUrl()) {
			// Validator must be bonded
			sender := msg.GetSigners()[0]
			validatorAddr := d.snapshotter.GetOperator(ctx, sender)
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

func msgRegistered(r cdctypes.InterfaceRegistry, targetURL string) bool {
	for _, url := range r.ListImplementations("axelarnet.v1beta1.Refundable") {
		if targetURL == url {
			return true
		}
	}
	return false
}
