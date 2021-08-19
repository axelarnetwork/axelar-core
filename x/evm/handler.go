package evm

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// NewHandler returns the handler of the EVM module
func NewHandler(k types.BaseKeeper, t types.TSS, v types.Voter, s types.Signer, n types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, t, n, s, v, snapshotter)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.LinkRequest:
			res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully linked {%s} and {%s}", res.DepositAddr, msg.RecipientAddr)
			}
			return result, err
		case *types.ConfirmChainRequest:
			res, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of EVM chain %s started", msg.Name)
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
				result.Log = fmt.Sprintf("votes on confirmation of transfer ownership %s started", msg.TxID.Hex())
			}
			return result, err
		case *types.VoteConfirmChainRequest:
			res, err := server.VoteConfirmChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = res.Log
			}
			return result, err
		case *types.VoteConfirmDepositRequest:
			res, err := server.VoteConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = res.Log
			}
			return result, err
		case *types.VoteConfirmTokenRequest:
			res, err := server.VoteConfirmToken(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = res.Log
			}
			return result, err
		case *types.VoteConfirmTransferKeyRequest:
			res, err := server.VoteConfirmTransferKey(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = res.Log
			}
			return result, err
		case *types.CreateDeployTokenRequest:
			res, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateBurnTokensRequest:
			res, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SignTxRequest:
			res, err := server.SignTx(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", res.TxID)
			}
			return result, err
		case *types.CreatePendingTransfersRequest:
			res, err := server.CreatePendingTransfers(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateTransferOwnershipRequest:
			res, err := server.CreateTransferOwnership(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateTransferOperatorshipRequest:
			res, err := server.CreateTransferOperatorship(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SignCommandsRequest:
			res, err := server.SignCommands(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully started signing batched commands with ID %s", hex.EncodeToString(res.BatchedCommandsID))
			}
			return result, err
		case *types.AddChainRequest:
			res, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully added new chain %s", msg.Name)

			}
			return result, err
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}
		k.Logger(ctx).Debug(res.Log)
		return res, nil
	}
}
