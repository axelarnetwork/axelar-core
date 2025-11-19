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
		signer     string
	)
	encodingConfig := app.MakeEncodingConfig()
	govAccount := rand.AccAddr()

	signerHasAnyRole := func() {
		signer = rand.AccAddr().String()
		permission.GetRoleFunc = func(sdk.Context, sdk.AccAddress) exported.Role {
			return exported.Role(rand.Of(maps.Keys(exported.Role_name)...))
		}
	}

	signerIsNot := func(role exported.Role) func() {
		return func() {
			signer = rand.AccAddr().String()
			permission.GetRoleFunc = func(sdk.Context, sdk.AccAddress) exported.Role {
				filtered := slices.Filter(maps.Keys(exported.Role_name), func(k int32) bool { return k != int32(role) })
				return exported.Role(rand.Of(filtered...))
			}
		}
	}

	signerIsGovAccount := func() {
		signer = govAccount.String()
		permission.GetRoleFunc = func(_ sdk.Context, addr sdk.AccAddress) exported.Role {
			if addr.Empty() {
				return exported.ROLE_UNRESTRICTED
			}
			return exported.ROLE_UNRESTRICTED
		}
	}

	signerIsEmpty := func() {
		signer = ""
		permission.GetRoleFunc = func(_ sdk.Context, addr sdk.AccAddress) exported.Role {
			if addr.Empty() {
				return exported.ROLE_UNRESTRICTED
			}
			panic("GetRole should not be called with non-empty address when signer is empty")
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

	msgRoleIsUnrestricted := func() { tx = txWithMsg(&evm.LinkRequest{Sender: signer}) }
	msgRoleIsUnspecified := func() { tx = txWithMsg(&banktypes.MsgSend{FromAddress: signer}) }
	msgRoleIsChainManagement := func() { tx = txWithMsg(&evm.CreateDeployTokenRequest{Sender: signer}) }
	msgRoleIsAccessControl := func() { tx = txWithMsg(&axelarnet.RegisterFeeCollectorRequest{Sender: signer}) }

	Given("a restricted tx ante handler", func() {
		permission = &mock.PermissionMock{}
		handler = ante.NewAnteHandlerDecorator(
			ante.ChainMessageAnteDecorators(ante.NewRestrictedTx(permission, encodingConfig.Codec)).ToAnteHandler())
	}).Branch(
		When("signer has any role", signerHasAnyRole).Branch(
			When("msg role is unrestricted", msgRoleIsUnrestricted).
				Then("let the msg through", letTxThrough),
			When("msg role is unspecified", msgRoleIsUnspecified).
				Then("let the msg through", letTxThrough),
		),

		When("signer is not chain management", signerIsNot(exported.ROLE_CHAIN_MANAGEMENT)).
			When("msg role is chain management", msgRoleIsChainManagement).
			Then("stop tx", stopTx),

		When("signer is not access control", signerIsNot(exported.ROLE_ACCESS_CONTROL)).
			When("msg role is access control", msgRoleIsAccessControl).
			Then("stop tx", stopTx),

		When("signer is empty", signerIsEmpty).Branch(
			When("msg role is unrestricted", msgRoleIsUnrestricted).
				Then("stop tx", stopTx),
			When("msg role is unspecified", msgRoleIsUnspecified).
				Then("stop tx", stopTx),
			When("msg role is chain management", msgRoleIsChainManagement).
				Then("stop tx", stopTx),
			When("msg role is access control", msgRoleIsAccessControl).
				Then("stop tx", stopTx),
		),

		// governance proposals bypass ante handlers, so the governance address should behave just like an account
		//without special permissions to prevent unintended permission check bypasses
		When("signer is gov account", signerIsGovAccount).Branch(
			When("msg role is chain management", msgRoleIsChainManagement).
				Then("stop tx", stopTx),
			When("msg role is access control", msgRoleIsAccessControl).
				Then("stop tx", stopTx),
		),
	).Run(t, 20)
}
