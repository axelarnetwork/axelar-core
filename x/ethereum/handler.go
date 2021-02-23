package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

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
			return handleMsgVoteVerifiedTx(ctx, k, v, n, msg)
		case types.MsgVerifyErc20TokenDeploy:
			return handleMsgVerifyErc20TokenDeploy(ctx, k, rpc, v, msg)
		case types.MsgVerifyErc20Deposit:
			return handleMsgVerifyErc20Deposit(ctx, k, rpc, v, msg)
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

func handleMsgVerifyErc20Deposit(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyErc20Deposit) (*sdk.Result, error) {
	txIDHex := msg.TxID.String()

	burnerInfo := k.GetBurnerInfo(ctx, msg.BurnerAddr)
	if burnerInfo == nil {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no burner info found for address %s", msg.BurnerAddr)
	}

	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txIDHex}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
	}

	erc20Deposit := types.Erc20Deposit{
		TxID:       msg.TxID,
		Amount:     msg.Amount,
		Symbol:     burnerInfo.Symbol,
		BurnerAddr: msg.BurnerAddr,
	}
	k.SetUnverifiedErc20Deposit(ctx, txIDHex, &erc20Deposit)

	txReceipt, blockNumber, err := getTransactionReceiptAndBlockNumber(ctx, rpc, msg.TxID)
	if err != nil {
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "cannot get transaction receipt %s or block number", txIDHex).Error())
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})

		return &sdk.Result{Log: err.Error()}, nil
	}

	tokenAddr := common.HexToAddress(burnerInfo.TokenAddr)
	if err := verifyErc20Deposit(ctx, k, txReceipt, blockNumber, msg.TxID, msg.Amount, msg.BurnerAddr, tokenAddr); err != nil {
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected erc20 deposit (%s) to burner address %s could not be verified", txIDHex, msg.BurnerAddr.String()).Error())
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})

		return &sdk.Result{Log: err.Error()}, nil
	}

	v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})

	return &sdk.Result{Log: fmt.Sprintf("successfully verified erc20 deposit %s", txIDHex)}, nil
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
		TokenAddr: tokenAddr.String(),
		Symbol:    msg.Symbol,
		Salt:      salt,
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

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot counter for key ID %s registered", keyID)
	}
	s, ok := snap.GetSnapshot(ctx, counter)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot found")
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
	txReceipt, blockNumber, err := getTransactionReceiptAndBlockNumber(ctx, rpc, tx.Hash())
	if err != nil {
		output := fmt.Sprintf("cannot get transaction receipt %s or block number: %v", txID, err)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		return &sdk.Result{Log: output}, nil
	}

	if !isTxFinalized(txReceipt, blockNumber, k.GetRequiredConfirmationHeight(ctx)) {
		output := fmt.Sprintf("expected transaction (%s) does not have enough confirmations", txID)
		k.Logger(ctx).Debug(output)
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		return &sdk.Result{Log: output}, nil

	}
	output := fmt.Sprintf("successfully verified transaction %s", txID)
	k.Logger(ctx).Debug(output)
	v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})
	return &sdk.Result{Log: output}, nil
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
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
		case types.MsgVerifyErc20Deposit{}.Type():
			txID := msg.PollMeta.ID
			k.ProcessVerificationErc20DepositResult(ctx, txID, confirmed.(bool))

			deposit := k.GetVerifiedErc20Deposit(ctx, txID)
			if deposit == nil {
				return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("erc20 deposit %s wasn't properly marked as verified", txID))
			}

			depositAddr := nexus.CrossChainAddress{Address: deposit.BurnerAddr.String(), Chain: exported.Ethereum}
			amount := sdk.NewInt64Coin(deposit.Symbol, deposit.Amount.BigInt().Int64())

			if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
				return nil, sdkerrors.Wrap(types.ErrEthereum, err.Error())
			}
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

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot counter for key ID %s registered", keyID)
	}
	s, ok := snap.GetSnapshot(ctx, counter)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot found")
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

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot counter for key ID %s registered", keyID)
	}
	s, ok := snap.GetSnapshot(ctx, counter)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrEthereum, "no snapshot found")
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

	txReceipt, blockNumber, err := getTransactionReceiptAndBlockNumber(ctx, rpc, msg.TxID)
	if err != nil {
		output := fmt.Sprintf("cannot get transaction receipt %s or block number: %v", msg.TxID.String(), err)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		return &sdk.Result{Log: output}, nil
	}

	if err := verifyERC20TokenDeploy(ctx, k, txReceipt, blockNumber, msg.Symbol, msg.GatewayAddr, tokenAddr); err != nil {
		output := fmt.Sprintf("expected erc20 token deploy (%s) could not be verified", msg.Symbol)
		k.Logger(ctx).Debug(sdkerrors.Wrap(err, output).Error())
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		return &sdk.Result{Log: output}, nil
	}

	output := fmt.Sprintf("successfully verified erc20 token deployment from transaction %s", msg.TxID.String())
	k.Logger(ctx).Debug(output)
	v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})
	return &sdk.Result{Log: output}, nil
}

func verifyERC20TokenDeploy(ctx sdk.Context, k keeper.Keeper, txReceipt *ethTypes.Receipt, blockNumber uint64, expectedSymbol string, gatewayAddr, expectedAddr common.Address) error {
	if !isTxFinalized(txReceipt, blockNumber, k.GetRequiredConfirmationHeight(ctx)) {
		return fmt.Errorf("transaction %s does not have enough confirmations yet", txReceipt.TxHash.String())
	}

	for _, log := range txReceipt.Logs {
		// Event is not emitted by the axelar gateway
		if log.Address != gatewayAddr {
			continue
		}

		// Event is not for a ERC20 token deployment
		symbol, tokenAddr, err := types.DecodeErc20TokenDeployEvent(log, k.GetERC20TokenDeploySignature(ctx))
		if err != nil {
			k.Logger(ctx).Debug(sdkerrors.Wrap(err, "event not for a a token deployment").Error())
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

func verifyErc20Deposit(ctx sdk.Context, k keeper.Keeper, txReceipt *ethTypes.Receipt, blockNumber uint64, txID common.Hash, amount sdk.Uint, burnerAddr common.Address, tokenAddr common.Address) error {
	if !isTxFinalized(txReceipt, blockNumber, k.GetRequiredConfirmationHeight(ctx)) {
		return fmt.Errorf("transaction %s does not have enough confirmations yet", txID.String())
	}

	actualAmount := sdk.ZeroUint()
	for _, log := range txReceipt.Logs {
		/* Event is not related to the token */
		if log.Address != tokenAddr {
			continue
		}

		_, to, transferAmount, err := types.DecodeErc20TransferEvent(log)
		/* Event is not an ERC20 transfer */
		if err != nil {
			continue
		}

		/* Transfer isn't sent to burner */
		if to != burnerAddr {
			continue
		}

		actualAmount = actualAmount.Add(transferAmount)
	}

	if !actualAmount.Equal(amount) {
		return fmt.Errorf("deposit amount in transaction %s for token %s and burner %s doesn not match what is expected", txID, tokenAddr.String(), burnerAddr.String())
	}

	return nil
}

func getTransactionReceiptAndBlockNumber(ctx sdk.Context, rpc types.RPCClient, txID common.Hash) (*ethTypes.Receipt, uint64, error) {
	txReceipt, err := rpc.TransactionReceipt(context.Background(), txID)
	if err != nil {
		return nil, 0, err
	}

	blockNumber, err := rpc.BlockNumber(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return txReceipt, blockNumber, nil
}

func isTxFinalized(txReceipt *ethTypes.Receipt, blockNumber uint64, confirmationHeight uint64) bool {
	return blockNumber-txReceipt.BlockNumber.Uint64() >= confirmationHeight
}
