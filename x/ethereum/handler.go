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
		case *types.MsgLink:
			return handleMsgLink(ctx, k, n, msg)
		case *types.MsgConfirmToken:
			return HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)
		case *types.MsgConfirmDeposit:
			return HandleMsgConfirmDeposit(ctx, k, v, s, msg)
		case *types.MsgVoteConfirmDeposit:
			return HandleMsgVoteConfirmDeposit(ctx, k, v, n, msg)
		case *types.MsgVoteConfirmToken:
			return HandleMsgVoteConfirmToken(ctx, k, v, n, msg)
		case *types.MsgSignDeployToken:
			return handleMsgSignDeployToken(ctx, k, s, snapshotter, v, msg)
		case *types.MsgSignBurnTokens:
			return handleMsgSignBurnTokens(ctx, k, s, snapshotter, v, msg)
		case *types.MsgSignTx:
			return handleMsgSignTx(ctx, k, s, snapshotter, v, msg)
		case *types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfers(ctx, k, s, n, snapshotter, v, msg)
		case types.MsgSignTransferOwnership:
			return handleMsgSignTransferOwnership(ctx, k, s, snapshotter, v, msg)
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

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, n types.Nexus, msg *types.MsgLink) (*sdk.Result, error) {
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
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.RecipientAddr),
		),
	)

	return &sdk.Result{
		Data:   []byte(burnerAddr.String()),
		Log:    logMsg,
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// HandleMsgConfirmTokenDeploy handles token deployment confirmation
func HandleMsgConfirmTokenDeploy(ctx sdk.Context, k types.EthKeeper, v types.Voter, signer types.Signer, n types.Nexus, msg *types.MsgConfirmToken) (*sdk.Result, error) {
	if n.IsAssetRegistered(ctx, exported.Ethereum.Name, msg.Symbol) {
		return nil, fmt.Errorf("token %s is already registered", msg.Symbol)
	}

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

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.TxID+"_"+msg.Symbol, ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
	if err := v.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	deploy := types.ERC20TokenDeploy{
		Symbol:    msg.Symbol,
		TokenAddr: tokenAddr.Hex(),
	}
	k.SetPendingTokenDeploy(ctx, poll, deploy)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxID),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, tokenAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeySymbol, msg.Symbol),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(k.GetRequiredConfirmationHeight(ctx), 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(poll))),
		),
	)

	return &sdk.Result{
		Log:    fmt.Sprintf("votes on confirmation of token deployment %s started", msg.TxID),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// HandleMsgConfirmDeposit handles deposit confirmations
func HandleMsgConfirmDeposit(ctx sdk.Context, k types.EthKeeper, v types.Voter, signer types.Signer, msg *types.MsgConfirmDeposit) (*sdk.Result, error) {
	_, state, ok := k.GetDeposit(ctx, msg.TxID, msg.BurnerAddr)
	switch {
	case !ok:
		break
	case state == types.CONFIRMED:
		return nil, fmt.Errorf("already confirmed")
	case state == types.BURNED:
		return nil, fmt.Errorf("already burned")
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

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.TxID+"_"+msg.BurnerAddr, ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
	if err := v.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	erc20Deposit := types.ERC20Deposit{
		TxID:       common.HexToHash(msg.TxID),
		Amount:     msg.Amount,
		Symbol:     burnerInfo.Symbol,
		BurnerAddr: msg.BurnerAddr,
	}
	k.SetPendingDeposit(ctx, poll, &erc20Deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
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
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// HandleMsgVoteConfirmDeposit handles votes for deposit confirmations
func HandleMsgVoteConfirmDeposit(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteConfirmDeposit) (*sdk.Result, error) {
	pendingDeposit, pollFound := k.GetPendingDeposit(ctx, msg.Poll)
	confirmedDeposit, state, depositFound := k.GetDeposit(ctx, msg.TxID, msg.BurnAddr)

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed token,
	// so we need to check that it matches the poll before deleting
	case depositFound && pollFound && confirmedDeposit == pendingDeposit:
		v.DeletePoll(ctx, msg.Poll)
		k.DeletePendingDeposit(ctx, msg.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case depositFound:
		switch state {
		case types.CONFIRMED:
			return &sdk.Result{Log: fmt.Sprintf("deposit in %s to address %s already confirmed", pendingDeposit.TxID, pendingDeposit.BurnerAddr)}, nil
		case types.BURNED:
			return &sdk.Result{Log: fmt.Sprintf("deposit in %s to address %s already spent", pendingDeposit.TxID, pendingDeposit.BurnerAddr)}, nil
		}
	case !pollFound:
		return nil, fmt.Errorf("no deposit found for poll %s", msg.Poll.String())
	case pendingDeposit.BurnerAddr != msg.BurnAddr || pendingDeposit.TxID.Hex() != msg.TxID:
		return nil, fmt.Errorf("deposit in %s to address %s does not match poll %s", msg.TxID, msg.BurnAddr, msg.Poll.String())
	default:
		// assert: the deposit is known and has not been confirmed before
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.Poll, msg.Confirmed); err != nil {
		return nil, err
	}

	result := v.Result(ctx, msg.Poll)
	if result == nil {
		return &sdk.Result{Log: fmt.Sprintf("not enough votes to confirm deposit in %s to %s yet", msg.TxID, msg.BurnAddr)}, nil
	}

	// assert: the poll has completed
	depositFound, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", msg.Poll.String(), result)
	}

	v.DeletePoll(ctx, msg.Poll)
	k.DeletePendingDeposit(ctx, msg.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeDepositConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.Poll))))

	if !depositFound {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &sdk.Result{
			Log:    fmt.Sprintf("deposit in %s to %s was discarded", msg.TxID, msg.BurnAddr),
			Events: ctx.EventManager().ABCIEvents(),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	depositAddr := nexus.CrossChainAddress{Address: pendingDeposit.BurnerAddr, Chain: exported.Ethereum}
	amount := sdk.NewInt64Coin(pendingDeposit.Symbol, pendingDeposit.Amount.BigInt().Int64())
	if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return nil, err
	}
	k.SetDeposit(ctx, pendingDeposit, types.CONFIRMED)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// HandleMsgVoteConfirmToken handles votes for token deployment confirmations
func HandleMsgVoteConfirmToken(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteConfirmToken) (*sdk.Result, error) {
	// is there an ongoing poll?
	token, pollFound := k.GetPendingTokenDeploy(ctx, msg.Poll)
	registered := n.IsAssetRegistered(ctx, exported.Ethereum.Name, msg.Symbol)
	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed token,
	// so we need to check that it matches the poll before deleting
	case registered && pollFound && token.Symbol == msg.Symbol:
		v.DeletePoll(ctx, msg.Poll)
		k.DeletePendingToken(ctx, msg.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case registered:
		return &sdk.Result{Log: fmt.Sprintf("token %s already confirmed", msg.Symbol)}, nil
	case !pollFound:
		return nil, fmt.Errorf("no token found for poll %s", msg.Poll.String())
	case token.Symbol != msg.Symbol:
		return nil, fmt.Errorf("token %s does not match poll %s", msg.Symbol, msg.Poll.String())
	default:
		// assert: the token is known and has not been confirmed before
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.Poll, msg.Confirmed); err != nil {
		return nil, err
	}

	result := v.Result(ctx, msg.Poll)
	if result == nil {
		return &sdk.Result{Log: fmt.Sprintf("not enough votes to confirm token %s yet", msg.Symbol)}, nil
	}

	// assert: the poll has completed
	confirmed, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", msg.Poll.String(), result)
	}

	v.DeletePoll(ctx, msg.Poll)
	k.DeletePendingToken(ctx, msg.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTokenConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.Poll))))

	if !confirmed {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &sdk.Result{
			Log:    fmt.Sprintf("token %s was discarded", msg.Symbol),
			Events: ctx.EventManager().ABCIEvents(),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	n.RegisterAsset(ctx, exported.Ethereum.Name, token.Symbol)

	return &sdk.Result{
		Log:    fmt.Sprintf("token %s deployment confirmed", token.Symbol),
		Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleMsgSignDeployToken(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg *types.MsgSignDeployToken) (*sdk.Result, error) {
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
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for deploy-token command %s", commandIDHex),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

func handleMsgSignBurnTokens(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg *types.MsgSignBurnTokens) (*sdk.Result, error) {
	deposits := k.GetConfirmedDeposits(ctx)

	if len(deposits) == 0 {
		return &sdk.Result{Log: fmt.Sprintf("no confirmed deposits found to burn")}, nil
	}

	chainID := k.GetNetwork(ctx).Params().ChainID

	var burnerInfos []types.BurnerInfo
	seen := map[string]bool{}
	for _, deposit := range deposits {
		if seen[deposit.BurnerAddr] {
			continue
		}
		burnerInfo := k.GetBurnerInfo(ctx, common.HexToAddress(deposit.BurnerAddr))
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", deposit.BurnerAddr)
		}
		burnerInfos = append(burnerInfos, *burnerInfo)
		seen[deposit.BurnerAddr] = true
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

	for _, deposit := range deposits {
		k.DeleteDeposit(ctx, deposit)
		k.SetDeposit(ctx, deposit, types.BURNED)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for burning %s token deposits, commandID: %s", exported.Ethereum.Name, commandIDHex),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

func handleMsgSignPendingTransfers(ctx sdk.Context, k keeper.Keeper, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter, v types.Voter, msg *types.MsgSignPendingTransfers) (*sdk.Result, error) {
	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Ethereum)

	if len(pendingTransfers) == 0 {
		return &sdk.Result{
			Log:    fmt.Sprintf("no pending transfer for chain %s found", exported.Ethereum.Name),
			Events: ctx.EventManager().ABCIEvents(),
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
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for %s pending transfers, commandID: %s", exported.Ethereum.Name, commandIDHex),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg *types.MsgSignTx) (*sdk.Result, error) {
	tx := msg.UnmarshaledTx()
	txID := tx.Hash().String()
	k.SetUnsignedTx(ctx, txID, tx)
	k.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := k.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txID, k.Codec().MustMarshalJSON(hash)))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
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
	// TODO: this is something that should be done after the signature has been successfully confirmed
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
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

func handleMsgSignTransferOwnership(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignTransferOwnership) (*sdk.Result, error) {

	chainID := k.GetNetwork(ctx).Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(msg.NewOwner))[:32])

	data, err := types.CreateTransferOwnershipCommandData(chainID, commandID, msg.NewOwner)
	if err != nil {
		return nil, err
	}

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	k.Logger(ctx).Info(fmt.Sprintf("storing data for transfer-ownership command %s", commandIDHex))
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &sdk.Result{
		Data:   commandID[:],
		Log:    fmt.Sprintf("successfully started signing protocol for transfer-ownership command %s", commandIDHex),
		Events: ctx.EventManager().Events(),
	}, nil
}
