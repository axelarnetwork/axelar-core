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

// ValidateWrappedMsgs rejects an authz MsgExec wrapping a role-restricted msg, and any
// authz MsgExec or BatchRequest wrapping another MsgExec or BatchRequest.
func ValidateWrappedMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		switch m := msg.(type) {
		case *authz.MsgExec:
			innerMsgs, err := m.GetMessages()
			if err != nil {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}
			gated, err := containsRoleGatedMsg(innerMsgs)
			if err != nil {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}
			if gated {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap role-restricted messages")
			}
			if containsNestingMsg(innerMsgs) {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "authz MsgExec must not wrap a BatchRequest or another MsgExec")
			}
		case *auxiliarytypes.BatchRequest:
			if containsNestingMsg(m.UnwrapMessages()) {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "BatchRequest must not wrap a BatchRequest or an authz MsgExec")
			}
		}
	}

	return nil
}

func containsRoleGatedMsg(msgs []sdk.Msg) (bool, error) {
	for _, msg := range msgs {
		role, err := permissionRole(msg)
		if err != nil {
			return false, err
		}
		switch role {
		case permission.ROLE_ACCESS_CONTROL, permission.ROLE_CHAIN_MANAGEMENT:
			return true, nil
		}
	}

	return false, nil
}

func containsNestingMsg(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *authz.MsgExec, *auxiliarytypes.BatchRequest:
			return true
		}
	}

	return false
}
