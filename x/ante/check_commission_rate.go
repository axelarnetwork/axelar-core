package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var minCommissionRate = sdk.NewDecWithPrec(5, 2)

// CheckCommissionRate checks if the validator commission rate is eligible
type CheckCommissionRate struct{}

// NewCheckCommissionRate constructor for CheckCommissionRate
func NewCheckCommissionRate() CheckCommissionRate {
	return CheckCommissionRate{}
}

// AnteHandle fails the transaction if it finds any validator registering a commission rate that is below the minimum
func (d CheckCommissionRate) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgCreateValidator:
			if msg.Commission.Rate.LT(minCommissionRate) || msg.Commission.MaxRate.LT(minCommissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator commission rate has to be >=%s", minCommissionRate.String())
			}
		case *stakingtypes.MsgEditValidator:
			if msg.CommissionRate != nil && msg.CommissionRate.LT(minCommissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator commission rate has to be >=%s", minCommissionRate.String())
			}
		default:
		}
	}

	return next(ctx, tx, simulate)
}
