package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		var signer sdk.AccAddress

		signers, _, err := d.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return ctx, sdkerrors.ErrInvalidRequest.Wrapf("failed to get signers for message %T: %s", msg, err)
		}
		if len(signers) != 0 {
			signer = signers[0]
		}

		role, err := permissionRole(msg)
		if err != nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}

		signerRole := d.permission.GetRole(ctx, signer)
		switch role {
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

func permissionRole(msg sdk.Msg) (permission.Role, error) {
	dm, ok := msg.(descriptor.Message)
	if !ok {
		return permission.ROLE_UNSPECIFIED, fmt.Errorf("message %T does not implement descriptor.Message", msg)
	}
	return permission.GetPermissionRole(dm), nil
}
