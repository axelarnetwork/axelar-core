package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestRestrictedTx(t *testing.T) {
	var (
		handler    sdk.AnteDecorator
		permission *mock.PermissionMock
		tx         *mock.TxMock
	)
	encodingConfig := app.MakeEncodingConfig()

	signerAnyRole := func() {
		permission.GetRoleFunc = func(sdk.Context, sdk.AccAddress) exported.Role {
			return exported.Role(rand.Of(maps.Keys(exported.Role_name)...))
		}
	}

	signerIsNot := func(role exported.Role) func() {
		return func() {
			permission.GetRoleFunc = func(sdk.Context, sdk.AccAddress) exported.Role {
				filtered := slices.Filter(maps.Keys(exported.Role_name), func(k int32) bool { return k != int32(role) })
				return exported.Role(rand.Of(filtered...))
			}
		}
	}

	noSigner := func() {
		permission.GetRoleFunc = func(_ sdk.Context, addr sdk.AccAddress) exported.Role {
			if len(addr) == 0 {
				return exported.ROLE_UNRESTRICTED
			}

			return exported.Role(rand.Of(maps.Keys(exported.Role_name)...))
		}
	}

	letTxThrough := func(t *testing.T) {
		_, err := handler.AnteHandle(sdk.Context{}, tx, false,
			func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.NoError(t, err)
	}

	stopTx := func(t *testing.T) {
		_, err := handler.AnteHandle(sdk.Context{}, tx, false,
			func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.Error(t, err)
	}

	txWithMsg := func(msg sdk.Msg) *mock.TxMock {
		return &mock.TxMock{
			GetMsgsFunc: func() []sdk.Msg {
				return slices.Expand(func(_ int) sdk.Msg {
					return msg
				}, int(rand.I64Between(1, 20)))
			},
		}
	}

	msgRoleIsUnrestricted := func() { tx = txWithMsg(&evm.LinkRequest{Sender: rand.AccAddr().String()}) }
	msgRoleIsUnspecified := func() { tx = txWithMsg(&banktypes.MsgSend{FromAddress: rand.AccAddr().String()}) }
	msgRoleIsChainManagement := func() { tx = txWithMsg(&evm.CreateDeployTokenRequest{Sender: rand.AccAddr().String()}) }
	msgRoleIsAccessControl := func() { tx = txWithMsg(&axelarnet.RegisterFeeCollectorRequest{Sender: rand.AccAddr().String()}) }

	Given("a restricted tx ante handler", func() {
		permission = &mock.PermissionMock{}
		handler = ante.NewAnteHandlerDecorator(
			ante.ChainMessageAnteDecorators(ante.NewRestrictedTx(permission, encodingConfig.Codec)).ToAnteHandler())
	}).Branch(
		When("msg role is unrestricted", msgRoleIsUnrestricted).
			When("signer has any role", signerAnyRole).
			Then("let the msg through", letTxThrough),

		When("msg role is unspecified", msgRoleIsUnspecified).
			When("signer has any role", signerAnyRole).
			Then("let the msg through", letTxThrough),

		When("msg role is chain management", msgRoleIsChainManagement).
			When("signer is not chain management", signerIsNot(exported.ROLE_CHAIN_MANAGEMENT)).
			Then("stop tx", stopTx),

		When("msg role is access control", msgRoleIsAccessControl).
			When("signer is not access control", signerIsNot(exported.ROLE_ACCESS_CONTROL)).
			Then("stop tx", stopTx),

		When("msg role is unrestricted", msgRoleIsUnrestricted).
			When("there is no signer", noSigner).
			Then("let the msg through", letTxThrough),

		When("msg role is unspecified", msgRoleIsUnspecified).
			When("there is no signer", noSigner).
			Then("let the msg through", letTxThrough),

		When("msg role is chain management", msgRoleIsChainManagement).
			When("there is no signer", noSigner).
			Then("stop tx", stopTx),

		When("msg role is access control", msgRoleIsAccessControl).
			When("there is no signer", noSigner).
			Then("stop tx", stopTx),
	).Run(t, 20)
}
