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
func (d CheckCommissionRate) AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error) {
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgCreateValidator:
			if msg.Commission.Rate.LT(minCommissionRate) || msg.Commission.MaxRate.LT(minCommissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator commission rate has to be >=%s", minCommissionRate.String())
			}
		case *stakingtypes.MsgEditValidator:
			// if commission rate isn't being changed, then let it pass
			if msg.CommissionRate == nil {
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

			// if existing commission rate is lower than the min rate, then let it pass.
			// if it's >= min rate, then don't allow decreasing it to < min rate.
			commissionRate := val.GetCommission()
			if commissionRate.GTE(minCommissionRate) && msg.CommissionRate.LT(minCommissionRate) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator commission rate has to be >=%s", minCommissionRate.String())
			}
		default:
		}
	}

	return next(ctx, msgs, simulate)
}
