package ante

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/protoc-gen-gogo/descriptor"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// RestrictedTx checks if the signer is authorized to send restricted transactions
type RestrictedTx struct {
	permission types.Permission
	cdc        codec.Codec
}

// NewRestrictedTx constructor for RestrictedTx
func NewRestrictedTx(permission types.Permission, cdc codec.Codec) RestrictedTx {
	return RestrictedTx{
		permission,
		cdc,
	}
}

// AnteHandle fails if the signer is not authorized to send the transaction
func (d RestrictedTx) AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error) {
	for _, msg := range msgs {
		if exec, ok := msg.(*authz.MsgExec); ok {
			if inner, err := exec.GetMessages(); err == nil && containsRoleGatedMsg(inner) {
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap role-restricted messages")
			}
		}

		var signer sdk.AccAddress

		signers, _, err := d.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return ctx, sdkerrors.ErrInvalidRequest.Wrapf("failed to get signers for message %T: %s", msg, err)
		}
		if len(signers) != 0 {
			signer = signers[0]
		}

		signerRole := d.permission.GetRole(ctx, signer)
		switch permissionRole(msg) {
		case permission.ROLE_ACCESS_CONTROL:
			if permission.ROLE_ACCESS_CONTROL != signerRole {
				return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "signer '%s' is not authorized to send transaction %T", signer, msg)
			}
		case permission.ROLE_CHAIN_MANAGEMENT:
			if permission.ROLE_CHAIN_MANAGEMENT != signerRole {
				return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "signer '%s' is not authorized to send transaction %T", signer, msg)
			}
		default:
			continue
		}
	}

	return next(ctx, msgs, simulate)
}

func permissionRole(msg sdk.Msg) permission.Role {
	dm, ok := msg.(descriptor.Message)
	if !ok {
		return permission.ROLE_UNSPECIFIED
	}
	return permission.GetPermissionRole(dm)
}

func containsRoleGatedMsg(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch permissionRole(msg) {
		case permission.ROLE_ACCESS_CONTROL, permission.ROLE_CHAIN_MANAGEMENT:
			return true
		}
	}

	return false
}
