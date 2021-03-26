package ethereum

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewHandler returns the handler of the ethereum module
func NewHandler(k keeper.Keeper, v types.Voter, s types.Signer, n types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return handleMsgLink(ctx, k, n, msg)
		case *types.MsgVoteConfirmation:
			return handleMsgVoteVerifiedTx(ctx, k, v, n, msg)
		case types.MsgVerifyErc20TokenDeploy:
			return handleMsgVerifyErc20TokenDeploy(ctx, k, v,s,  msg)
		case types.MsgVerifyErc20Deposit:
			return handleMsgVerifyErc20Deposit(ctx, k, v,s, msg)
		case types.MsgSignDeployToken:
			return handleMsgSignDeployToken(ctx, k, s, snapshotter, v, msg)
		case types.MsgSignBurnTokens:
			return handleMsgSignBurnTokens(ctx, k, s, snapshotter, v, msg)
		case types.MsgSignTx:
			return handleMsgSignTx(ctx, k, s, snapshotter, v, msg)
		case types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfers(ctx, k, s, n, snapshotter, v, msg)
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

func handleMsgVerifyErc20Deposit(ctx sdk.Context, k keeper.Keeper, v types.Voter, signer types.Signer, msg types.MsgVerifyErc20Deposit) (*sdk.Result, error) {
	if res := k.GetVerifiedErc20Deposit(ctx, msg.TxID, msg.BurnerAddr); res != nil {
		return nil, fmt.Errorf("already verified")
	}
	if res := k.GetArchivedErc20Deposit(ctx, msg.TxID, msg.BurnerAddr); res != nil {
		return nil, fmt.Errorf("already spent")
	}

	burnerInfo := k.GetBurnerInfo(ctx, common.HexToAddress(msg.BurnerAddr))
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", msg.BurnerAddr)
	}

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.Type(), msg.TxID+msg.BurnerAddr, ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
	if err := v.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	erc20Deposit := types.ERC20Deposit{
		TxID:       common.HexToHash(msg.TxID),
		Amount:     msg.Amount,
		Symbol:     burnerInfo.Symbol,
		BurnerAddr: msg.BurnerAddr,
	}
	k.SetUnconfirmedERC20Deposit(ctx, poll, &erc20Deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxID),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, msg.BurnerAddr),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddr),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(k.GetRequiredConfirmationHeight(ctx), 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(poll))),
		),
	)

	return &sdk.Result{
		Log:    fmt.Sprintf("votes on confirmation of deposit %s started", msg.TxID),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyErc20TokenDeploy(ctx sdk.Context, k keeper.Keeper, v types.Voter, signer types.Signer,msg types.MsgVerifyErc20TokenDeploy) (*sdk.Result, error) {
	gatewayAddr, ok := k.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.Type(), msg.TxID+msg.Symbol, ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
	if err := v.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	deploy := types.ERC20TokenDeploy{
		TxID:      common.HexToHash(msg.TxID),
		Symbol:    msg.Symbol,
		TokenAddr: tokenAddr.Hex(),
	}
	k.SetUnverifiedErc20TokenDeploy(ctx, &deploy)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxID),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, tokenAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(k.GetRequiredConfirmationHeight(ctx), 10)),
			sdk.NewAttribute(types.AttributeKeyDeploySig, k.GetERC20TokenDeploySignature(ctx).Hex()),
			sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(poll))),
		),
	)

	return &sdk.Result{
		Log:    fmt.Sprintf("votes on confirmation of token deployment %s started", msg.TxID),
		Events: ctx.EventManager().Events(),
	}, nil
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
			sdk.NewAttribute(types.AttributeKeyBurnAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.RecipientAddr),
		),
	)

	return &sdk.Result{
		Data:   []byte(burnerAddr.String()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignPendingTransfers(ctx sdk.Context, k keeper.Keeper, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
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

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = signer.StartSign(ctx, v, keyID, commandIDHex, signHash.Bytes(), snapshot)
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
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteConfirmation) (*sdk.Result, error) {
	if token := k.GetVerifiedToken(ctx, msg.PollMeta.ID); token != nil {
		return &sdk.Result{Log: fmt.Sprintf("token %s already verified", token.Symbol)}, nil
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.PollMeta, msg.Confirmed); err != nil {
		return nil, err
	}

	if result := v.Result(ctx, msg.PollMeta); result != nil {
		event := sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.PollMeta))),
			sdk.NewAttribute(types.AttributeKeyResult, strconv.FormatBool(result.(bool))))

		switch msg.PollMeta.Type {
		case types.MsgVerifyErc20TokenDeploy{}.Type():
			k.ProcessVerificationTokenResult(ctx, msg.PollMeta.ID, result.(bool))

			token := k.GetVerifiedToken(ctx, msg.PollMeta.ID)
			if token == nil {
				k.Logger(ctx).Info(fmt.Sprintf("poll %s could not be verified by the validators", msg.PollMeta.ID))
				break
			}

			n.RegisterAsset(ctx, exported.Ethereum.Name, token.Symbol)
			event = event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeKeyActionToken),
				sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.PollMeta))),
				sdk.NewAttribute(types.AttributeKeyTxID, common.Bytes2Hex(token.TxID[:])))

		case types.MsgVerifyErc20Deposit{}.Type():
			k.ProcessVerificationErc20DepositResult(ctx, msg.PollMeta.ID, result.(bool))

			deposit := k.GetVerifiedErc20Deposit(ctx, msg.PollMeta.ID, "")
			if deposit == nil {
				k.Logger(ctx).Info(fmt.Sprintf("poll %s could not be verified by the validators", msg.PollMeta.ID))
				break
			}

			depositAddr := nexus.CrossChainAddress{Address: deposit.BurnerAddr, Chain: exported.Ethereum}
			amount := sdk.NewInt64Coin(deposit.Symbol, deposit.Amount.BigInt().Int64())
			if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
				return nil, err
			}
			event = event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeKeyActionDeposit),
				sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.PollMeta))),
				sdk.NewAttribute(types.AttributeKeyTxID, common.Bytes2Hex(deposit.TxID[:])))

		default:
			k.Logger(ctx).Debug(fmt.Sprintf("unknown verification message type: %s", msg.PollMeta.Type))
			event = event.AppendAttributes(
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeKeyActionUnknown),
				sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.PollMeta))),
			)
		}

		ctx.EventManager().EmitEvent(event)
		v.DeletePoll(ctx, msg.PollMeta)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignDeployToken(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignDeployToken) (*sdk.Result, error) {
	chainID := k.GetParams(ctx).Network.Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(msg.TokenName))[:32])

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, msg.TokenName, msg.Symbol, msg.Decimals, msg.Capacity)
	if err != nil {
		return nil, err
	}

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = signer.StartSign(ctx, v, keyID, commandIDHex, signHash.Bytes(), snapshot)
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

func handleMsgSignBurnTokens(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignBurnTokens) (*sdk.Result, error) {
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

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for burn command %s", commandIDHex))
	k.SetCommandData(ctx, commandID, data)

	k.Logger(ctx).Info(fmt.Sprintf("signing burn command [%s] for token deposits to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = signer.StartSign(ctx, v, keyID, commandIDHex, signHash.Bytes(), snapshot)
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

func getUniqueBurnerAddrs(deposits []types.ERC20Deposit) []common.Address {
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

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignTx) (*sdk.Result, error) {
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
			sdk.NewAttribute(types.AttributeKeyTxID, txID),
		),
	)

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = signer.StartSign(ctx, v, keyID, txID, hash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	// TODO: this is something that should be done after the signature has been successfully verified
	if tx.To() == nil && bytes.Equal(tx.Data(), k.GetGatewayByteCodes(ctx)) {

		pub, ok := signer.GetCurrentKey(ctx, exported.Ethereum, tss.MasterKey)
		if !ok {
			return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub.Value), tx.Nonce())
		k.SetGatewayAddress(ctx, addr)
	}

	return &sdk.Result{
		Data:   []byte(txID),
		Log:    fmt.Sprintf("successfully started signing protocol for transaction with ID %s.", txID),
		Events: ctx.EventManager().Events(),
	}, nil
}
