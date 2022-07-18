package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/utils/slices"
)

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HandlerDecorator is an ante decorator wrapper for an ante handler
type HandlerDecorator struct {
	handler sdk.AnteHandler
}

// NewAnteHandlerDecorator constructor for HandlerDecorator
func NewAnteHandlerDecorator(handler sdk.AnteHandler) HandlerDecorator {
	return HandlerDecorator{handler}
}

// AnteHandle wraps the next AnteHandler to perform custom pre- and post-processing
func (decorator HandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if newCtx, err = decorator.handler(ctx, tx, simulate); err != nil {
		return newCtx, err
	}

	return next(newCtx, tx, simulate)
}

// LogMsgDecorator logs all messages in blocks
type LogMsgDecorator struct {
	cdc codec.Codec
}

// NewLogMsgDecorator is the constructor for LogMsgDecorator
func NewLogMsgDecorator(cdc codec.Codec) LogMsgDecorator {
	return LogMsgDecorator{cdc: cdc}
}

// AnteHandle logs all messages in blocks
func (d LogMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate || ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		logger(ctx).Debug(fmt.Sprintf("received message of type %s in block %d: %s",
			proto.MessageName(msg),
			ctx.BlockHeight(),
			string(d.cdc.MustMarshalJSON(msg)),
		))
	}

	return next(ctx, tx, simulate)
}

// UndelegateDecorator checks if the unbonding validator holds any multiSig share of active crypto keys
type UndelegateDecorator struct {
	multiSig    types.MultiSig
	nexus       types.Nexus
	snapshotter types.Snapshotter
}

// NewUndelegateDecorator constructor for UndelegateDecorator
func NewUndelegateDecorator(multiSig types.MultiSig, nexus types.Nexus, snapshotter types.Snapshotter) UndelegateDecorator {
	return UndelegateDecorator{
		multiSig,
		nexus,
		snapshotter,
	}
}

// AnteHandle fails the transaction if it finds any validator holding multiSig share of active keys is trying to unbond
func (d UndelegateDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgUndelegate:
			valAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}

			// only restrict a validator from unbonding it's self-delegation
			if !delegatorAddress.Equals(valAddress) {
				continue
			}

			chains := d.nexus.GetChains(ctx)

			for _, chain := range chains {
				nextKeyID, idFound := d.multiSig.GetNextKeyID(ctx, chain.Name)
				key, keyFound := d.multiSig.GetKey(ctx, nextKeyID)
				if !(idFound && keyFound && holdsShares(key, valAddress)) {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding multiSig share of %s's next key %s", valAddress, chain.Name, nextKeyID)
				}

				activeKeyIDs := d.multiSig.GetActiveKeyIDs(ctx, chain.Name)
				for _, activeKeyID := range activeKeyIDs {
					key, keyFound := d.multiSig.GetKey(ctx, activeKeyID)
					if !(keyFound && holdsShares(key, valAddress)) {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding multiSig share of %s's active key %s", valAddress, chain.Name, activeKeyID)
					}
				}
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}

func holdsShares(key exported.Key, valAddress sdk.ValAddress) bool {
	return slices.Any(key.GetParticipants(), func(v sdk.ValAddress) bool { return v.Equals(valAddress) })
}
