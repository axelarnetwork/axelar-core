package axelarnet

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// NewHandler returns the handler of the Cosmos module
func NewHandler(k keeper.Keeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, ibcK keeper.IBCKeeper) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, n, b, a, ibcK)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.LinkRequest:
			res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully linked deposit %s to recipient %s", res.DepositAddr, msg.RecipientAddr)
			}
			return result, err
		case *types.ConfirmDepositRequest:
			res, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully confirmed deposit to {%s}", msg.DepositAddress.String())
			}
			return result, err
		case *types.ExecutePendingTransfersRequest:
			res, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = "successfully executed pending transfers"
			}
			return result, err
		case *types.AddCosmosBasedChainRequest:
			res, err := server.AddCosmosBasedChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully added chain %s", msg.CosmosChain)
			}
			return result, err
		case *types.RegisterAssetRequest:
			res, err := server.RegisterAsset(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully registered asset %s to chain %s", msg.Asset.Denom, msg.Chain)
			}
			return result, err
		case *types.RouteIBCTransfersRequest:
			res, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = "successfully routed IBC transfers"
			}
			return result, err
		case *types.RegisterFeeCollectorRequest:
			res, err := server.RegisterFeeCollector(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		case *types.RetryIBCTransferRequest:
			res, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		case *types.RouteMessageRequest:
			res, err := server.RouteMessage(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		case *types.CallContractRequest:
			res, err := server.CallContract(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully enqueued contract call for contract %s on chain %s", msg.ContractAddress, msg.Chain)
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
			return nil, sdkerrors.Wrap(types.ErrAxelarnet, err.Error())
		}
		return res, nil
	}
}

// NewProposalHandler returns the handler for the proposals of the axelarnet module
func NewProposalHandler(k keeper.Keeper, nexusK types.Nexus, accountK types.AccountKeeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CallContractsProposal:
			for _, contractCall := range c.ContractCalls {
				sender := nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: accountK.GetModuleAddress(govtypes.ModuleName).String()}
				recipient := nexus.CrossChainAddress{Chain: funcs.MustOk(nexusK.GetChain(ctx, contractCall.Chain)), Address: contractCall.ContractAddress}
				// axelar gateway expects keccak256 hashes for payloads
				payloadHash := crypto.Keccak256(contractCall.Payload)

				msgID, txID, nonce := nexusK.GenerateMessageID(ctx)
				msg := nexus.NewGeneralMessage(msgID, sender, recipient, payloadHash, txID, nonce, nil)

				events.Emit(ctx, &types.ContractCallSubmitted{
					MessageID:        msg.ID,
					Sender:           msg.GetSourceAddress(),
					SourceChain:      msg.GetSourceChain(),
					DestinationChain: msg.GetDestinationChain(),
					ContractAddress:  msg.GetDestinationAddress(),
					PayloadHash:      msg.PayloadHash,
					Payload:          contractCall.Payload,
				})

				if err := nexusK.SetNewMessage(ctx, msg); err != nil {
					return sdkerrors.Wrap(err, "failed to add general message")
				}

				k.Logger(ctx).Debug(fmt.Sprintf("successfully enqueued contract call for contract address %s on chain %s from sender %s with message id %s", recipient.Address, recipient.Chain.String(), sender.Address, msg.ID),
					types.AttributeKeyDestinationChain, recipient.Chain.String(),
					types.AttributeKeyDestinationAddress, recipient.Address,
					types.AttributeKeySourceAddress, sender.Address,
					types.AttributeKeyMessageID, msg.ID,
					types.AttributeKeyPayloadHash, hex.EncodeToString(payloadHash),
				)
			}

			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized axelarnet proposal content type: %T", c)
		}
	}
}
