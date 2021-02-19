package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewHandler returns the handler of the ethereum module
func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer, snap snapshot.Snapshotter, n types.Nexus) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return handleMsgLink(ctx, k, n, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, rpc, v, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, msg)
		case types.MsgVerifyErc20TokenDeploy:
			return handleMsgVerifyErc20TokenDeploy(ctx, k, rpc, v, msg)
		case types.MsgSignDeployToken:
			return handleMsgSignDeployToken(ctx, k, s, snap, msg)
		case types.MsgSignTx:
			return handleMsgSignTx(ctx, k, s, snap, msg)
		case types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfersTx(ctx, k, s, snap, n, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, n types.Nexus, msg types.MsgLink) (*sdk.Result, error) {
	gatewayAddr := common.HexToAddress(msg.GatewayAddr)
	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, gatewayAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	burnerAddr, salt, err := k.GetBurnerAddressAndSalt(ctx, tokenAddr, msg.RecipientAddr, gatewayAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	senderChain, ok := n.GetChain(ctx, exported.Ethereum.Name)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "%s is not a registered chain", exported.Ethereum.Name)
	}
	recipientChain, ok := n.GetChain(ctx, msg.RecipientChain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "unknown recipient chain")
	}
	n.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddr.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddr})

	burnerInfo := types.BurnerInfo{
		Symbol: msg.Symbol,
		Salt:   salt,
	}
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)

	logMsg := fmt.Sprintf("successfully linked {%s} and {%s}", burnerAddr.String(), msg.RecipientAddr)
	k.Logger(ctx).Info(logMsg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.RecipientAddr),
		),
	)

	return &sdk.Result{
		Data:   []byte(burnerAddr.String()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignPendingTransfersTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, n types.Nexus, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Ethereum)

	if len(pendingTransfers) == 0 {
		return &sdk.Result{
			Data:   nil,
			Log:    fmt.Sprintf("no pending transfer for chain %s found", exported.Ethereum.Name),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	chainID := k.GetNetwork(ctx).Params().ChainID

	data, err := types.CreateMintCommandData(chainID, pendingTransfers)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", exported.Ethereum.Name)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for %s pending transfers, commandID: %s", exported.Ethereum.Name, commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying ethereum transaction")
	tx := msg.UnmarshaledTx()
	txID := tx.Hash().String()

	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txID}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
		),
	)

	k.SetUnverifiedTx(ctx, txID, tx)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/
	receipt, err := rpc.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		output := fmt.Sprintf("could not retrieve Ethereum receipt for transaction %s: %v", txID, err)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		return recordVote(ctx, k, poll, false, output, v), nil
	}
	if err = verifyTx(ctx, k, rpc, receipt.BlockNumber.Uint64()); err != nil {
		output := fmt.Sprintf("expected transaction (%s) could not be verified: %v", txID, err)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		return recordVote(ctx, k, poll, false, output, v), nil

	}
	output := fmt.Sprintf("successfully verified transaction %s", txID)
	k.Logger(ctx).Debug(output)
	return recordVote(ctx, k, poll, true, output, v), nil
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeVotingData, strconv.FormatBool(msg.VotingData)),
	)

	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {

		switch msg.PollMeta.Type {
		case types.MsgVerifyTx{}.Type():
			k.ProcessVerificationTxResult(ctx, msg.PollMeta.ID, confirmed.(bool))
		case types.MsgVerifyErc20TokenDeploy{}.Type():
			k.ProcessVerificationTokenResult(ctx, msg.PollMeta.ID, confirmed.(bool))
		default:
			k.Logger(ctx).Debug(fmt.Sprintf("unknown verification message type: %s", msg.PollMeta.Type))
		}

		v.DeletePoll(ctx, msg.Poll())
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollConfirmed, strconv.FormatBool(confirmed.(bool))))
	}

	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignDeployToken(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, msg types.MsgSignDeployToken) (*sdk.Result, error) {
	chainID := k.GetParams(ctx).Network.Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(msg.TokenName))[:32])

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, msg.TokenName, msg.Symbol, msg.Decimals, msg.Capacity)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", exported.Ethereum.Name)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	k.SaveTokenInfo(ctx, msg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for deploy-token command %s", commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap snapshot.Snapshotter, msg types.MsgSignTx) (*sdk.Result, error) {
	tx := msg.UnmarshaledTx()
	txID := tx.Hash().String()
	k.SetRawTx(ctx, txID, tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txID, k.Codec().MustMarshalJSON(hash)))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
		),
	)

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no master key for chain %s found", exported.Ethereum.Name)
	}

	s, ok := snap.GetLatestSnapshot(ctx)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrEthereum, "no snapshot found")
	}
	err = signer.StartSign(ctx, keyID, txID, hash.Bytes(), s.Validators)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	return &sdk.Result{
		Data:   []byte(txID),
		Log:    fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", txID),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyErc20TokenDeploy(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyErc20TokenDeploy) (*sdk.Result, error) {

	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, msg.GatewayAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.TxID.String()}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, msg.TxID.String()),
		),
	)

	deploy := types.Erc20TokenDeploy{
		TxID:      msg.TxID,
		Symbol:    msg.Symbol,
		TokenAddr: tokenAddr,
	}
	k.SetUnverifiedErc20TokenDeploy(ctx, &deploy)

	receipt, err := rpc.TransactionReceipt(context.Background(), msg.TxID)
	if err != nil {
		output := fmt.Sprintf("could not retrieve Ethereum receipt for transaction %s: %v", msg.TxID.String(), err)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		return recordVote(ctx, k, poll, false, output, v), nil
	}
	err = verifyTx(ctx, k, rpc, receipt.BlockNumber.Uint64())
	if err != nil {
		output := fmt.Sprintf("transaction '%s' could not be verified", msg.TxID.String())
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		return recordVote(ctx, k, poll, false, output, v), nil
	}
	if err := verifyERC20TokenDeploy(ctx, receipt, k, msg.Symbol, msg.GatewayAddr, tokenAddr); err != nil {
		output := fmt.Sprintf("expected erc20 token deploy (%s) could not be verified", msg.Symbol)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		return recordVote(ctx, k, poll, false, output, v), nil
	}

	output := fmt.Sprintf("successfully verified erc20 token deployment from transaction %s", msg.TxID.String())
	k.Logger(ctx).Debug(output)
	return recordVote(ctx, k, poll, true, output, v), nil
}

func verifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, blockNum uint64) error {

	height, err := rpc.BlockNumber(context.Background())
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Ethereum block number")
	}

	if (height - blockNum) < k.GetRequiredConfirmationHeight(ctx) {
		return fmt.Errorf("not enough confirmations yet")
	}
	return nil
}

func verifyERC20TokenDeploy(ctx sdk.Context, receipt *ethTypes.Receipt, keeper keeper.Keeper, expectedSymbol string, gatewayAddr, expectedAddr common.Address) error {
	for _, log := range receipt.Logs {
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// Event is not for a ERC20 token deployment
		symbol, tokenAddr, err := types.DecodeErc20TokenDeployEvent(log, keeper.GetERC20TokenDeploySignature(ctx))
		if err != nil {
			keeper.Logger(ctx).Debug(sdkerrors.Wrap(err, "event not for a a token deployment").Error())
			continue
		}

		// Symbol does not match
		if symbol != expectedSymbol {
			continue
		}

		// token address does not match
		if tokenAddr != expectedAddr {
			continue
		}

		// if we reach this point, it means that the log matches what we want to verify,
		// so the function can return with no error
		return nil
	}

	return fmt.Errorf("failed to verify token deployment for symbol '%s' at contract address '%s'", expectedSymbol, expectedAddr.String())
}

func recordVote(ctx sdk.Context, keeper keeper.Keeper, poll vote.PollMeta, vote bool, msg string, voter vote.Voter) *sdk.Result {
	if err := voter.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: vote}); err != nil {
		keeper.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
		return &sdk.Result{
			Log:    err.Error(),
			Data:   keeper.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}
	}
	return &sdk.Result{
		Log:    msg,
		Data:   keeper.Codec().MustMarshalBinaryLengthPrefixed(vote),
		Events: ctx.EventManager().Events(),
	}
}
