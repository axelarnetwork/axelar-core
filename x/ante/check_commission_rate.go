package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
)

var minCommissionRate = sdk.NewDecWithPrec(5, 2)

// CheckCommissionRate checks if the validator commission rate is eligible
type CheckCommissionRate struct {
	staking types.Staking
}

// NewCheckCommissionRate constructor for CheckCommissionRate
func NewCheckCommissionRate(staking types.Staking) CheckCommissionRate {
	return CheckCommissionRate{
		staking,
	}
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
			if msg.CommissionRate == nil || msg.CommissionRate.GTE(minCommissionRate) {
				continue
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return ctx, err
			}

			val := d.staking.Validator(ctx, valAddr)
			if val == nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "not a validator")
			}

			commissionRate := val.GetCommission()
			if commissionRate.GTE(minCommissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator commission rate has to be >=%s", minCommissionRate.String())
			}

			if commissionRate.LT(minCommissionRate) && msg.CommissionRate.LT(commissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator with existing commission rate below min %s cannot be decreased more", minCommissionRate.String())
			}
		default:
		}
	}

	return next(ctx, tx, simulate)
}
