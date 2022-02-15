package ante

import (
	"fmt"

	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
)

// CheckRefundFeeDecorator record potential refund for tss and vote txs
type CheckRefundFeeDecorator struct {
	registry    cdctypes.InterfaceRegistry
	ak          antetypes.AccountKeeper
	staking     types.Staking
	reward      types.Reward
	snapshotter types.Snapshotter
}

// NewCheckRefundFeeDecorator constructor for CheckRefundFeeDecorator
func NewCheckRefundFeeDecorator(registry cdctypes.InterfaceRegistry, ak antetypes.AccountKeeper, staking types.Staking, snapshotter types.Snapshotter, reward types.Reward) CheckRefundFeeDecorator {
	return CheckRefundFeeDecorator{
		registry,
		ak,
		staking,
		reward,
		snapshotter,
	}
}

// AnteHandle record qualified refund for the tss and vote transactions
func (d CheckRefundFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	if !anyRefundable(msgs) {
		return next(ctx, tx, simulate)
	}

	if err := d.qualifyForRefund(ctx, msgs); err != nil {
		return ctx, err
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	fee := feeTx.GetFee()
	if len(fee) > 0 {
		req := msgs[0].(*rewardtypes.RefundMsgRequest)
		err := d.reward.SetPendingRefund(ctx, *req, fee[0])
		if err != nil {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	return next(ctx, tx, simulate)
}

func anyRefundable(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}

	for _, msg := range msgs {
		switch msg.(type) {
		case *rewardtypes.RefundMsgRequest:
			return true
		}
	}
	return false
}

func (d CheckRefundFeeDecorator) qualifyForRefund(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *rewardtypes.RefundMsgRequest:
			if !msgRegistered(d.registry, msg.InnerMessage.TypeUrl) {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("message type %s is not refundable", msg.InnerMessage.TypeUrl))
			}

			sender := msg.GetSigners()[0]
			validatorAddr := d.snapshotter.GetOperator(ctx, sender)
			if validatorAddr == nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "signer does not belong to a validator")
			}

			validator := d.staking.Validator(ctx, validatorAddr)
			if validator == nil || !validator.IsBonded() {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "associated validator is not bonded")
			}
		default:
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("message type %T is not refundable", msg))
		}
	}

	return nil
}

func msgRegistered(r cdctypes.InterfaceRegistry, targetURL string) bool {
	for _, url := range r.ListImplementations("reward.v1beta1.Refundable") {
		if targetURL == url {
			return true
		}
	}
	return false
}
