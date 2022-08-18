package evm

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// NewHandler returns the handler of the EVM module
func NewHandler(k types.BaseKeeper, v types.Voter, n types.Nexus, snapshotter types.Snapshotter, staking types.StakingKeeper, slashing types.SlashingKeeper, multisigKeeper types.MultisigKeeper) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, n, v, snapshotter, staking, slashing, multisigKeeper)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.SetGatewayRequest:
			res, err := server.SetGateway(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.LinkRequest:
			res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully linked deposit %s to recipient %s", res.DepositAddr, msg.RecipientAddr)
			}
			return result, err
		case *types.ConfirmTokenRequest:
			res, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of token deployment %s started", msg.TxID.Hex())
			}
			return result, err
		case *types.ConfirmDepositRequest:
			res, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of deposit %s started", msg.TxID.Hex())
			}
			return result, err
		case *types.ConfirmTransferKeyRequest:
			res, err := server.ConfirmTransferKey(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of transfer operatorship %s started", msg.TxID.Hex())
			}
			return result, err
		case *types.ConfirmGatewayTxRequest:
			res, err := server.ConfirmGatewayTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.CreateDeployTokenRequest:
			res, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateBurnTokensRequest:
			res, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreatePendingTransfersRequest:
			res, err := server.CreatePendingTransfers(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateTransferOperatorshipRequest:
			res, err := server.CreateTransferOperatorship(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SignCommandsRequest:
			res, err := server.SignCommands(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				if res.CommandCount == 0 {
					result.Log = "no command to sign found"
				} else {
					result.Log = fmt.Sprintf("successfully started signing batched commands with ID %s", hex.EncodeToString(res.BatchedCommandsID))
				}
			}
			return result, err
		case *types.AddChainRequest:
			res, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully added new chain %s", msg.Name)
			}
			return result, err
		case *types.RetryFailedEventRequest:
			res, err := server.RetryFailedEvent(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			// if the error is not already a registered error the error message would be obscured, so wrap it in a general registered error
			if !utils.IsABCIError(err) {
				err = sdkerrors.Wrap(types.ErrEVM, err.Error())
			}
			k.Logger(ctx).Debug(err.Error())
			return nil, err
		}
		if len(res.Log) > 0 {
			k.Logger(ctx).Debug(res.Log)
		}
		return res, nil
	}
}
