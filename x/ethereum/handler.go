package ethereum

import (
	"bytes"
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
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewHandler returns the handler of the ethereum module
func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer, n types.Nexus) sdk.Handler {
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return handleMsgLink(ctx, k, n, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, n, msg)
		case types.MsgVerifyErc20TokenDeploy:
			return handleMsgVerifyErc20TokenDeploy(ctx, k, rpc, v, msg)
		case types.MsgVerifyErc20Deposit:
			return handleMsgVerifyErc20Deposit(ctx, k, rpc, v, msg)
		case types.MsgSignDeployToken:
			return handleMsgSignDeployToken(ctx, k, s, msg)
		case types.MsgSignBurnTokens:
			return handleMsgSignBurnTokens(ctx, k, s, msg)
		case types.MsgSignTx:
			return handleMsgSignTx(ctx, k, s, msg)
		case types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfers(ctx, k, s, n, msg)
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

func handleMsgVerifyErc20Deposit(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyErc20Deposit) (*sdk.Result, error) {
	txID := common.BytesToHash(msg.TxID[:])
	txIDHex := txID.String()

	burnerInfo := k.GetBurnerInfo(ctx, common.HexToAddress(msg.BurnerAddr))
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", msg.BurnerAddr)
	}

	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txIDHex}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

	erc20Deposit := types.Erc20Deposit{
		TxID:       msg.TxID,
		Amount:     msg.Amount,
		Symbol:     burnerInfo.Symbol,
		BurnerAddr: msg.BurnerAddr,
	}
	k.SetUnverifiedErc20Deposit(ctx, txIDHex, &erc20Deposit)

	txReceipt, blockNumber, err := getTransactionReceiptAndBlockNumber(rpc, txID)
	if err != nil {
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})

		return &sdk.Result{Log: sdkerrors.Wrapf(err, "cannot get transaction receipt %s or block number", txIDHex).Error()}, nil
	}

	if err := verifyErc20Deposit(ctx, k, txReceipt, blockNumber, txID, msg.Amount, common.HexToAddress(msg.BurnerAddr), common.HexToAddress(burnerInfo.TokenAddr)); err != nil {
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		log := sdkerrors.Wrapf(err, "expected erc20 deposit (%s) to burner address %s could not be verified", txIDHex, msg.BurnerAddr).Error()
		return &sdk.Result{Log: log}, nil
	}

	v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})

	return &sdk.Result{Log: fmt.Sprintf("successfully verified erc20 deposit %s", txIDHex)}, nil
}

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, n types.Nexus, msg types.MsgLink) (*sdk.Result, error) {
	gatewayAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	senderChain, ok := n.GetChain(ctx, exported.Ethereum.Name)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", exported.Ethereum.Name)
	}
	recipientChain, ok := n.GetChain(ctx, msg.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	found := n.IsAssetRegistered(ctx, recipientChain.Name, msg.Symbol)
	if !found {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", exported.Ethereum.NativeAsset, recipientChain.Name)
	}

	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	burnerAddr, salt, err := k.GetBurnerAddressAndSalt(ctx, tokenAddr, msg.RecipientAddr, gatewayAddr)
	if err != nil {
		return nil, err
	}

	n.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddr.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddr})

	var array [common.HashLength]byte
	copy(array[:], salt.Bytes())
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.Hex(),
		Symbol:    msg.Symbol,
		Salt:      array,
	}
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)

	logMsg := fmt.Sprintf("successfully linked {%s} and {%s}", burnerAddr.String(), msg.RecipientAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeBurnAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.RecipientAddr),
		),
	)

	return &sdk.Result{
		Data:   []byte(burnerAddr.String()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignPendingTransfers(ctx sdk.Context, k keeper.Keeper, signer types.Signer, n types.Nexus, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Ethereum)

	if len(pendingTransfers) == 0 {
		return &sdk.Result{
			Log:    fmt.Sprintf("no pending transfer for chain %s found", exported.Ethereum.Name),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	chainID := k.GetNetwork(ctx).Params().ChainID

	data, err := types.CreateMintCommandData(chainID, pendingTransfers)
	if err != nil {
		return nil, err
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes())
	if err != nil {
		return nil, err
	}

	// TODO: Archive pending transfers after signing is completed
	for _, pendingTransfer := range pendingTransfers {
		n.ArchivePendingTransfer(ctx, pendingTransfer)
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

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	txID := msg.PollMeta.ID
	if token := k.GetVerifiedToken(ctx, txID); token != nil {
		return &sdk.Result{Log: fmt.Sprintf("token %s already verified", token.Symbol)}, nil
	}

	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	eventType := types.EventTypeUnknownVerificationResult
	if result := v.Result(ctx, msg.Poll()); result != nil {
		switch msg.PollMeta.Type {
		case types.MsgVerifyErc20TokenDeploy{}.Type():
			k.ProcessVerificationTokenResult(ctx, txID, result.(bool))

			token := k.GetVerifiedToken(ctx, txID)
			if token == nil {
				return nil, fmt.Errorf("token %s wasn't properly marked as verified", txID)
			}

			n.RegisterAsset(ctx, exported.Ethereum.Name, token.Symbol)
			eventType = types.EventTypeDepositVerificationResult

		case types.MsgVerifyErc20Deposit{}.Type():
			k.ProcessVerificationErc20DepositResult(ctx, txID, result.(bool))

			deposit := k.GetVerifiedErc20Deposit(ctx, txID)
			if deposit == nil {
				return nil, fmt.Errorf("erc20 deposit %s wasn't properly marked as verified", txID)
			}

			depositAddr := nexus.CrossChainAddress{Address: deposit.BurnerAddr, Chain: exported.Ethereum}
			amount := sdk.NewInt64Coin(deposit.Symbol, deposit.Amount.BigInt().Int64())
			if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
				return nil, err
			}
			eventType = types.EventTypeTokenVerificationResult

		default:
			k.Logger(ctx).Debug(fmt.Sprintf("unknown verification message type: %s", msg.PollMeta.Type))
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(eventType,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyTxID, txID),
				sdk.NewAttribute(types.AttributeKeyResult, strconv.FormatBool(result.(bool)))))

		v.DeletePoll(ctx, msg.Poll())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignDeployToken(ctx sdk.Context, k keeper.Keeper, signer types.Signer, msg types.MsgSignDeployToken) (*sdk.Result, error) {
	chainID := k.GetParams(ctx).Network.Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(msg.TokenName))[:32])

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, msg.TokenName, msg.Symbol, msg.Decimals, msg.Capacity)
	if err != nil {
		return nil, err
	}

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes())
	if err != nil {
		return nil, err
	}

	k.SetTokenInfo(ctx, msg)

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

func handleMsgSignBurnTokens(ctx sdk.Context, k keeper.Keeper, signer types.Signer, msg types.MsgSignBurnTokens) (*sdk.Result, error) {
	deposits := k.GetVerifiedErc20Deposits(ctx)

	if len(deposits) == 0 {
		return &sdk.Result{Log: fmt.Sprintf("no verified token deposits found to burn")}, nil
	}

	chainID := k.GetNetwork(ctx).Params().ChainID
	burnerAddrs := getUniqueBurnerAddrs(deposits)

	var burnerInfos []types.BurnerInfo
	for _, burnerAddr := range burnerAddrs {
		burnerInfo := k.GetBurnerInfo(ctx, burnerAddr)
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", burnerAddr)
		}

		burnerInfos = append(burnerInfos, *burnerInfo)
	}

	data, err := types.CreateBurnCommandData(chainID, ctx.BlockHeight(), burnerInfos)
	if err != nil {
		return nil, err
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := signer.GetCurrentMasterKeyID(ctx, exported.Ethereum)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for burn command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing burn command [%s] for token deposits to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	err = signer.StartSign(ctx, keyID, commandIDHex, signHash.Bytes())
	if err != nil {
		return nil, err
	}

	// TODO: Archive token deposits after signing is completed
	for _, deposit := range deposits {
		k.ArchiveErc20Deposit(ctx, common.BytesToHash(deposit.TxID[:]).String())
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
		Log:    fmt.Sprintf("successfully started signing protocol for burning %s token deposits, commandID: %s", exported.Ethereum.Name, commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}

func getUniqueBurnerAddrs(deposits []types.Erc20Deposit) []common.Address {
	var burnerAddrs []common.Address
	burnerAddrSeen := map[common.Address]bool{}

	for _, deposit := range deposits {
		burnerAddr := common.HexToAddress(deposit.BurnerAddr)
		if burnerAddrSeen[burnerAddr] {
			continue
		}

		burnerAddrSeen[burnerAddr] = true
		burnerAddrs = append(burnerAddrs, burnerAddr)
	}

	return burnerAddrs
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, msg types.MsgSignTx) (*sdk.Result, error) {
	tx := msg.UnmarshaledTx()
	txID := tx.Hash().String()
	k.SetRawTx(ctx, txID, tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	err = signer.StartSign(ctx, keyID, txID, hash.Bytes())
	if err != nil {
		return nil, err
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	// TODO: this is something that should be done after the signature has been successfully verified
	if tx.To() == nil && bytes.Equal(tx.Data(), k.GetGatewayByteCodes(ctx)) {

		pub, ok := signer.GetCurrentMasterKey(ctx, exported.Ethereum)
		if !ok {
			return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub), tx.Nonce())
		k.SetGatewayAddress(ctx, addr)
	}

	return &sdk.Result{
		Data:   []byte(txID),
		Log:    fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", txID),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyErc20TokenDeploy(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyErc20TokenDeploy) (*sdk.Result, error) {
	txID := common.BytesToHash(msg.TxID[:])
	txHex := txID.String()

	if res := k.GetVerifiedErc20Deposit(ctx, txHex); res != nil {
		return nil, fmt.Errorf("already verified")
	}
	if res := k.GetArchivedErc20Deposit(ctx, txHex); res != nil {
		return nil, fmt.Errorf("already spent")
	}

	gatewayAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: txHex}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txHex),
		),
	)

	deploy := types.Erc20TokenDeploy{
		TxID:      msg.TxID,
		Symbol:    msg.Symbol,
		TokenAddr: tokenAddr.Hex(),
	}
	k.SetUnverifiedErc20TokenDeploy(ctx, &deploy)

	txReceipt, blockNumber, err := getTransactionReceiptAndBlockNumber(rpc, txID)
	if err != nil {
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		output := sdkerrors.Wrapf(err, "cannot get transaction receipt %s or block number", txHex).Error()
		return &sdk.Result{Log: output}, nil
	}

	if err := verifyERC20TokenDeploy(ctx, k, txReceipt, blockNumber, msg.Symbol, gatewayAddr, tokenAddr); err != nil {
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		output := sdkerrors.Wrapf(err, "expected erc20 token deploy (%s) could not be verified", msg.Symbol).Error()
		return &sdk.Result{Log: output}, nil
	}

	v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})
	output := fmt.Sprintf("successfully verified erc20 token deployment from transaction %s", txHex)
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

func getTransactionReceiptAndBlockNumber(rpc types.RPCClient, txID common.Hash) (*ethTypes.Receipt, uint64, error) {
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
	return blockNumber-txReceipt.BlockNumber.Uint64()+1 >= confirmationHeight
}
