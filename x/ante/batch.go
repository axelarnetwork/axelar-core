package ante

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// txWithUnwrappedMsgs implements the FeeTx interface
type txWithUnwrappedMsgs struct {
	sdk.FeeTx
	messages []sdk.Msg
}

func newTxWithUnwrappedMsgs(tx sdk.Tx) (txWithUnwrappedMsgs, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return txWithUnwrappedMsgs{}, errorsmod.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	unpacked, err := unpackMsgs(tx.GetMsgs())
	if err != nil {
		return txWithUnwrappedMsgs{}, err
	}

	return txWithUnwrappedMsgs{feeTx, unpacked}, nil
}

func (t txWithUnwrappedMsgs) GetMsgs() []sdk.Msg {
	return t.messages
}

// BatchDecorator unwraps batch requests and passes them to the next AnteHandler
type BatchDecorator struct {
	cdc codec.Codec
}

// NewBatchDecorator is the constructor for BatchDecorator
func NewBatchDecorator(cdc codec.Codec) BatchDecorator {
	return BatchDecorator{
		cdc,
	}
}

// AnteHandle unwraps batch requests and passes them to the next AnteHandler
func (b BatchDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := ValidateWrappedMsgs(tx.GetMsgs()); err != nil {
		return ctx, err
	}

	tx, err := newTxWithUnwrappedMsgs(tx)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func unpackMsgs(msgs []sdk.Msg) ([]sdk.Msg, error) {
	var unpackedMsgs []sdk.Msg

	for _, msg := range msgs {
		unpackedMsgs = append(unpackedMsgs, msg)

		switch m := msg.(type) {
		case *auxiliarytypes.BatchRequest:
			inner, err := unpackMsgs(m.UnwrapMessages())
			if err != nil {
				return nil, err
			}
			unpackedMsgs = append(unpackedMsgs, inner...)
		case *authz.MsgExec:
			innerMsgs, err := m.GetMessages()
			if err != nil {
				return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}
			unpackedMsgs = append(unpackedMsgs, innerMsgs...)
		}
	}

	return unpackedMsgs, nil
}

// ValidateWrappedMsgs rejects unsupported wrapper contents: an authz MsgExec must not
// wrap a role-restricted message, another MsgExec, or a batch request, and a batch
// request must not wrap a MsgExec. It is the sole gate preventing role-restricted
// messages from being delegated via authz, so it must run at every ante entry point.
func ValidateWrappedMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		switch m := msg.(type) {
		case *authz.MsgExec:
			innerMsgs, err := m.GetMessages()
			if err != nil {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}
			if containsRoleGatedMsg(innerMsgs) {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap role-restricted messages")
			}
			for _, innerMsg := range innerMsgs {
				switch innerMsg.(type) {
				case *authz.MsgExec:
					return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap another MsgExec")
				case *auxiliarytypes.BatchRequest:
					return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap a batch request")
				}
			}
		case *auxiliarytypes.BatchRequest:
			innerMsgs := m.UnwrapMessages()
			for _, innerMsg := range innerMsgs {
				if _, ok := innerMsg.(*authz.MsgExec); ok {
					return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "batch request must not wrap an authz MsgExec")
				}
			}
			if err := ValidateWrappedMsgs(innerMsgs); err != nil {
				return err
			}
		}
	}

	return nil
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
