package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestRestrictedTx(t *testing.T) {
	var (
		handler    ante.RestrictedTx
		permission *mock.PermissionMock
		tx         *mock.TxMock
	)

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

	txWithMsg := func(msg descriptor.Message) *mock.TxMock {
		return &mock.TxMock{
			GetMsgsFunc: func() []sdk.Msg {
				return slices.Expand(func(_ int) sdk.Msg {
					return &mock.MsgMock{
						GetSignersFunc: func() []sdk.AccAddress {
							return slices.Expand(func(_ int) sdk.AccAddress { return rand.AccAddr() }, int(rand.I64Between(1, 5)))
						},
						DescriptorFunc: msg.Descriptor,
					}
				}, int(rand.I64Between(1, 20)))
			},
		}
	}

	msgRoleIsUnrestricted := func() { tx = txWithMsg(&evm.LinkRequest{}) }
	msgRoleIsUnspecified := func() { tx = txWithMsg(&banktypes.MsgSend{}) }
	msgRoleIsChainManagement := func() { tx = txWithMsg(&tss.RotateKeyRequest{}) }
	msgRoleIsAccessControl := func() { tx = txWithMsg(&axelarnet.RegisterFeeCollectorRequest{}) }

	Given("a restricted tx ante handler", func() {
		permission = &mock.PermissionMock{}
		handler = ante.NewRestrictedTx(permission)
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
	).Run(t, 20)
}
