package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestCheckCommissionRate(t *testing.T) {
	var (
		handler ante.CheckCommissionRate
		tx      *mock.TxMock
		msg     sdk.Msg
	)

	givenCheckCommissionRateAnteHandler := Given("the check commission rate ante handler", func() {
		handler = ante.NewCheckCommissionRate()
	})

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgCreateValidator with commission rate below minimum is received", func() {
			msg = &stakingtypes.MsgCreateValidator{
				Commission: stakingtypes.CommissionRates{
					Rate:    sdk.NewDecWithPrec(49, 3),
					MaxRate: sdk.NewDecWithPrec(51, 3),
				},
			}

			tx = &mock.TxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{msg}
				},
			}
		}).
		Then("should return an error", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.Error(t, err)
		}).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgCreateValidator with max commission rate below minimum is received", func() {
			msg = &stakingtypes.MsgCreateValidator{
				Commission: stakingtypes.CommissionRates{
					Rate:    sdk.NewDecWithPrec(51, 3),
					MaxRate: sdk.NewDecWithPrec(49, 3),
				},
			}

			tx = &mock.TxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{msg}
				},
			}
		}).
		Then("should return an error", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.Error(t, err)
		}).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgEditValidator with commission rate below minimum is received", func() {
			commissionRate := sdk.NewDecWithPrec(49, 3)
			msg = &stakingtypes.MsgEditValidator{
				CommissionRate: &commissionRate,
			}

			tx = &mock.TxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{msg}
				},
			}
		}).
		Then("should return an error", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.Error(t, err)
		}).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with MsgEditValidator with unspecified commission rate is received", func() {
			msg = &stakingtypes.MsgEditValidator{
				CommissionRate: nil,
			}

			tx = &mock.TxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{msg}
				},
			}
		}).
		Then("should go through", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.NoError(t, err)
		}).
		Run(t)

	givenCheckCommissionRateAnteHandler.
		When("a tx with eligible MsgCreateValidator and MsgEditValidator is received", func() {
			commissionRate := sdk.NewDecWithPrec(51, 3)

			tx = &mock.TxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{
						&stakingtypes.MsgEditValidator{
							CommissionRate: &commissionRate,
						},
						&stakingtypes.MsgCreateValidator{
							Commission: stakingtypes.CommissionRates{
								Rate:    commissionRate,
								MaxRate: commissionRate,
							},
						},
					}
				},
			}
		}).
		Then("should go through", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false, func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.NoError(t, err)
		}).
		Run(t)
}
