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
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestRestrictedTx(t *testing.T) {
	var (
		handler    ante.RestrictedTx
		permission *mock.PermissionMock
		msgs       []sdk.Msg
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

	letMsgsThrough := func(t *testing.T) {
		_, err := handler.AnteHandle(sdk.Context{}, msgs, false,
			func(sdk.Context, []sdk.Msg, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.NoError(t, err)
	}

	stopMsgs := func(t *testing.T) {
		_, err := handler.AnteHandle(sdk.Context{}, msgs, false,
			func(sdk.Context, []sdk.Msg, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		assert.Error(t, err)
	}

	toSdkMsgs := func(msg descriptor.Message) []sdk.Msg {
		return slices.Expand(func(_ int) sdk.Msg {
			return &mock.MsgMock{
				GetSignersFunc: func() []sdk.AccAddress {
					return slices.Expand(func(_ int) sdk.AccAddress { return rand.AccAddr() }, int(rand.I64Between(1, 5)))
				},
				DescriptorFunc: msg.Descriptor,
			}
		}, int(rand.I64Between(1, 20)))
	}

	msgRoleIsUnrestricted := func() { msgs = toSdkMsgs(&evm.LinkRequest{}) }
	msgRoleIsUnspecified := func() { msgs = toSdkMsgs(&banktypes.MsgSend{}) }
	msgRoleIsChainManagement := func() { msgs = toSdkMsgs(&evm.CreateDeployTokenRequest{}) }
	msgRoleIsAccessControl := func() { msgs = toSdkMsgs(&axelarnet.RegisterFeeCollectorRequest{}) }

	Given("a restricted tx ante handler", func() {
		permission = &mock.PermissionMock{}
		handler = ante.NewRestrictedTx(permission)
	}).Branch(
		When("msg role is unrestricted", msgRoleIsUnrestricted).
			When("signer has any role", signerAnyRole).
			Then("let the msg through", letMsgsThrough),

		When("msg role is unspecified", msgRoleIsUnspecified).
			When("signer has any role", signerAnyRole).
			Then("let the msg through", letMsgsThrough),

		When("msg role is chain management", msgRoleIsChainManagement).
			When("signer is not chain management", signerIsNot(exported.ROLE_CHAIN_MANAGEMENT)).
			Then("stop tx", stopMsgs),

		When("msg role is access control", msgRoleIsAccessControl).
			When("signer is not access control", signerIsNot(exported.ROLE_ACCESS_CONTROL)).
			Then("stop tx", stopMsgs),
	).Run(t, 20)
}
