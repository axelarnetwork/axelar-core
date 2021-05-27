package ethereum

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// NewHandler returns the handler of the ethereum module
func NewHandler(k keeper.Keeper, v types.Voter, s types.Signer, n types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, n, s, v, snapshotter)
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
		case *types.ConfirmTokenRequest:
			res, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of token deployment %s started", msg.TxID)
			}
			return result, err
		case *types.ConfirmDepositRequest:
			res, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("votes on confirmation of deposit %s started", msg.TxID)
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
		case *types.SignDeployTokenRequest:
			res, err := server.SignDeployToken(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully started signing protocol for deploy-token command %s", hex.EncodeToString(res.CommandID))
			}
			return result, err
		case *types.SignBurnTokensRequest:
			res, err := server.SignBurnTokens(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				if res.CommandID == nil {
					result.Log = "no confirmed deposits found to burn"
				} else {
					result.Log = fmt.Sprintf("successfully started signing protocol for burning %s token deposits, commandID: %s",
						exported.Ethereum.Name, hex.EncodeToString(res.CommandID))
				}
			}
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SignTxRequest:
			res, err := server.SignTx(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", res.TxID)
			}
			return result, err
		case *types.SignPendingTransfersRequest:
			res, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				if res.CommandID == nil {
					result.Log = fmt.Sprintf("no pending transfer for chain %s found", exported.Ethereum.Name)
				} else {
					result.Log = fmt.Sprintf("successfully started signing protocol for %s pending transfers, commandID: %s",
						exported.Ethereum.Name, hex.EncodeToString(res.CommandID))
				}
			}
			return result, err
		case *types.SignTransferOwnershipRequest:
			res, err := server.SignTransferOwnership(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully started signing protocol for transfer-ownership command %s", hex.EncodeToString(res.CommandID))
			}
			return result, err
		case *types.AddChainRequest:

			res, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = res.Log
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
			return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
		}
		k.Logger(ctx).Debug(res.Log)
		return res, nil
	}
}
