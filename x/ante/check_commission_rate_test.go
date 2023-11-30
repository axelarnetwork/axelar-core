package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestCheckCommissionRate(t *testing.T) {
	var (
		handler sdk.AnteDecorator
		tx      *mock.TxMock
		msg     sdk.Msg
		staking *mock.StakingMock
	)

	valAddr := rand.ValAddr().String()

	letTxThrough := func(t *testing.T) {
		tx = &mock.TxMock{
			GetMsgsFunc: func() []sdk.Msg {
				return []sdk.Msg{msg}
			},
		}

		_, err := handler.AnteHandle(sdk.Context{}, tx, false,
			func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.NoError(t, err)
	}

	stopTx := func(t *testing.T) {
		tx = &mock.TxMock{
			GetMsgsFunc: func() []sdk.Msg {
				return []sdk.Msg{msg}
			},
		}

		_, err := handler.AnteHandle(sdk.Context{}, tx, false,
			func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.Error(t, err)
	}

	setValidatorCommission := func(commission sdk.Dec) {
		staking.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
			return stakingtypes.Validator{
				Commission: stakingtypes.NewCommission(commission, sdk.OneDec(), sdk.NewDecWithPrec(10, 3)),
			}
		}
	}

	createValidator := func(commissionRate sdk.Dec, maxCommissionRate sdk.Dec) func() {
		return func() {
			msg = &stakingtypes.MsgCreateValidator{
				Commission: stakingtypes.CommissionRates{
					Rate:    commissionRate,
					MaxRate: maxCommissionRate,
				},
			}
		}
	}

	editValidator := func(commissionRate sdk.Dec, newCommissionRate sdk.Dec) func() {
		return func() {
			msg = &stakingtypes.MsgEditValidator{
				ValidatorAddress: valAddr,
				CommissionRate:   &newCommissionRate,
			}
			setValidatorCommission(commissionRate)
		}
	}

	givenCheckCommissionRateAnteHandler := Given("the check commission rate ante handler", func() {
		staking = &mock.StakingMock{}
		handler = ante.NewAnteHandlerDecorator(
			ante.ChainMessageAnteDecorators(ante.NewCheckCommissionRate(staking)).ToAnteHandler())
	})

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgCreateValidator with commission rate below minimum is received", createValidator(sdk.NewDecWithPrec(49, 3), sdk.NewDecWithPrec(51, 3))).
		Then("should return an error", stopTx).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgCreateValidator with max commission rate below minimum is received", createValidator(sdk.NewDecWithPrec(51, 3), sdk.NewDecWithPrec(49, 3))).
		Then("should return an error", stopTx).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgEditValidator with commission rate below minimum is received", editValidator(sdk.NewDecWithPrec(50, 3), sdk.NewDecWithPrec(49, 3))).
		Then("should return an error", stopTx).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgEditValidator for validator with existing commission rate below minimum being increased is received", editValidator(sdk.NewDecWithPrec(39, 3), sdk.NewDecWithPrec(49, 3))).
		Then("should go through", letTxThrough).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgEditValidator with unspecified commission rate is received", func() {
			msg = &stakingtypes.MsgEditValidator{
				ValidatorAddress: valAddr,
				CommissionRate:   nil,
			}
			setValidatorCommission(sdk.NewDecWithPrec(49, 3))
		}).
		Then("should go through", letTxThrough).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with eligible MsgCreateValidator is received", createValidator(sdk.NewDecWithPrec(50, 3), sdk.NewDecWithPrec(51, 3))).
		Then("should go through", letTxThrough).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with eligible MsgEditValidator is received", editValidator(sdk.NewDecWithPrec(50, 3), sdk.NewDecWithPrec(51, 3))).
		Then("should go through", letTxThrough).
		Run(t)
}
