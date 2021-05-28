package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.EthKeeper
	signer      types.Signer
	nexus       types.Nexus
	voter       types.Voter
	snapshotter types.Snapshotter
}

// NewMsgServerImpl returns an implementation of the bitcoin MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.EthKeeper, n types.Nexus, s types.Signer, v types.Voter, snap types.Snapshotter) types.MsgServiceServer {
	return msgServer{
		EthKeeper:   keeper,
		signer:      s,
		nexus:       n,
		voter:       v,
		snapshotter: snap,
	}
}

func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	gatewayAddr, ok := s.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	senderChain, ok := s.nexus.GetChain(ctx, exported.Ethereum.Name)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", exported.Ethereum.Name)
	}
	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	found := s.nexus.IsAssetRegistered(ctx, recipientChain.Name, req.Symbol)
	if !found {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", exported.Ethereum.NativeAsset, recipientChain.Name)
	}

	tokenAddr, err := s.GetTokenAddress(ctx, req.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	burnerAddr, salt, err := s.GetBurnerAddressAndSalt(ctx, tokenAddr, req.RecipientAddr, gatewayAddr)
	if err != nil {
		return nil, err
	}

	s.nexus.LinkAddresses(ctx,
		nexus.CrossChainAddress{Chain: senderChain, Address: burnerAddr.String()},
		nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr})

	burnerInfo := types.BurnerInfo{
		TokenAddress: types.Address(tokenAddr),
		Symbol:       req.Symbol,
		Salt:         types.Hash(salt),
	}
	s.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, burnerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyAddress, req.RecipientAddr),
		),
	)

	return &types.LinkResponse{DepositAddr: burnerAddr.Hex()}, nil
}

// ConfirmToken handles token deployment confirmation
func (s msgServer) ConfirmToken(c context.Context, req *types.ConfirmTokenRequest) (*types.ConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if s.nexus.IsAssetRegistered(ctx, exported.Ethereum.Name, req.Symbol) {
		return nil, fmt.Errorf("token %s is already registered", req.Symbol)
	}

	gatewayAddr, ok := s.GetGatewayAddress(ctx)
	if !ok {
		return nil, fmt.Errorf("axelar gateway address not set")
	}

	tokenAddr, err := s.GetTokenAddress(ctx, req.Symbol, gatewayAddr)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	poll := vote.NewPollMeta(types.ModuleName, req.TxID.Hex()+"_"+req.Symbol)
	if err := s.voter.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	deploy := types.ERC20TokenDeployment{
		Symbol:       req.Symbol,
		TokenAddress: types.Address(tokenAddr),
	}
	s.SetPendingTokenDeployment(ctx, poll, deploy)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyGatewayAddress, gatewayAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, tokenAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeySymbol, req.Symbol),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(s.GetRequiredConfirmationHeight(ctx), 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&poll))),
		),
	)

	return &types.ConfirmTokenResponse{}, nil
}

// ConfirmDeposit handles deposit confirmations
func (s msgServer) ConfirmDeposit(c context.Context, req *types.ConfirmDepositRequest) (*types.ConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_, state, ok := s.GetDeposit(ctx, common.Hash(req.TxID), common.Address(req.BurnerAddress))
	switch {
	case !ok:
		break
	case state == types.CONFIRMED:
		return nil, fmt.Errorf("already confirmed")
	case state == types.BURNED:
		return nil, fmt.Errorf("already burned")
	}

	burnerInfo := s.GetBurnerInfo(ctx, common.Address(req.BurnerAddress))
	if burnerInfo == nil {
		return nil, fmt.Errorf("no burner info found for address %s", req.BurnerAddress)
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	poll := vote.NewPollMeta(types.ModuleName, req.TxID.Hex()+"_"+req.BurnerAddress.Hex())
	if err := s.voter.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}

	erc20Deposit := types.ERC20Deposit{
		TxID:          types.Hash(req.TxID),
		Amount:        req.Amount,
		Symbol:        burnerInfo.Symbol,
		BurnerAddress: req.BurnerAddress,
	}
	s.SetPendingDeposit(ctx, poll, &erc20Deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
			sdk.NewAttribute(types.AttributeKeyTxID, req.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyAmount, req.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyBurnAddress, req.BurnerAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(s.GetRequiredConfirmationHeight(ctx), 10)),
			sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&poll))),
		),
	)

	return &types.ConfirmDepositResponse{}, nil
}

// VoteConfirmDeposit handles votes for deposit confirmations
func (s msgServer) VoteConfirmDeposit(c context.Context, req *types.VoteConfirmDepositRequest) (*types.VoteConfirmDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pendingDeposit, pollFound := s.GetPendingDeposit(ctx, req.Poll)
	confirmedDeposit, state, depositFound := s.GetDeposit(ctx, common.Hash(req.TxID), common.Address(req.BurnAddress))

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed token,
	// so we need to check that it matches the poll before deleting
	case depositFound && pollFound && confirmedDeposit == pendingDeposit:
		s.voter.DeletePoll(ctx, req.Poll)
		s.DeletePendingDeposit(ctx, req.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case depositFound:
		switch state {
		case types.CONFIRMED:
			return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("deposit in %s to address %s already confirmed", pendingDeposit.TxID.Hex(), pendingDeposit.BurnerAddress.Hex())}, nil
		case types.BURNED:
			return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("deposit in %s to address %s already spent", pendingDeposit.TxID.Hex(), pendingDeposit.BurnerAddress.Hex())}, nil
		}
	case !pollFound:
		return nil, fmt.Errorf("no deposit found for poll %s", req.Poll.String())
	case pendingDeposit.BurnerAddress != req.BurnAddress || pendingDeposit.TxID != req.TxID:
		return nil, fmt.Errorf("deposit in %s to address %s does not match poll %s", req.TxID, req.BurnAddress.Hex(), req.Poll.String())
	default:
		// assert: the deposit is known and has not been confirmed before
	}

	if err := s.voter.TallyVote(ctx, req.Sender, req.Poll, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	result := s.voter.Result(ctx, req.Poll)
	if result == nil {
		return &types.VoteConfirmDepositResponse{Log: fmt.Sprintf("not enough votes to confirm deposit in %s to %s yet", req.TxID, req.BurnAddress.Hex())}, nil
	}

	// assert: the poll has completed
	confirmed, ok := result.(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.Poll.String(), result)
	}

	s.Logger(ctx).Info(fmt.Sprintf("ethereum deposit confirmation result is %s", result))
	s.voter.DeletePoll(ctx, req.Poll)
	s.DeletePendingDeposit(ctx, req.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeDepositConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.Poll))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmDepositResponse{
			Log: fmt.Sprintf("deposit in %s to %s was discarded", req.TxID, req.BurnAddress.Hex()),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	depositAddr := nexus.CrossChainAddress{Address: pendingDeposit.BurnerAddress.Hex(), Chain: exported.Ethereum}
	amount := sdk.NewInt64Coin(pendingDeposit.Symbol, pendingDeposit.Amount.BigInt().Int64())
	if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return nil, err
	}
	s.SetDeposit(ctx, pendingDeposit, types.CONFIRMED)

	return &types.VoteConfirmDepositResponse{}, nil
}

// VoteConfirmToken handles votes for token deployment confirmations
func (s msgServer) VoteConfirmToken(c context.Context, req *types.VoteConfirmTokenRequest) (*types.VoteConfirmTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	// is there an ongoing poll?
	token, pollFound := s.GetPendingTokenDeployment(ctx, req.Poll)
	registered := s.nexus.IsAssetRegistered(ctx, exported.Ethereum.Name, req.Symbol)
	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed token,
	// so we need to check that it matches the poll before deleting
	case registered && pollFound && token.Symbol == req.Symbol:
		s.voter.DeletePoll(ctx, req.Poll)
		s.DeletePendingToken(ctx, req.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case registered:
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("token %s already confirmed", req.Symbol)}, nil
	case !pollFound:
		return nil, fmt.Errorf("no token found for poll %s", req.Poll.String())
	case token.Symbol != req.Symbol:
		return nil, fmt.Errorf("token %s does not match poll %s", req.Symbol, req.Poll.String())
	default:
		// assert: the token is known and has not been confirmed before
	}

	if err := s.voter.TallyVote(ctx, req.Sender, req.Poll, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	result := s.voter.Result(ctx, req.Poll)
	if result == nil {
		return &types.VoteConfirmTokenResponse{Log: fmt.Sprintf("not enough votes to confirm token %s yet", req.Symbol)}, nil
	}

	// assert: the poll has completed
	confirmed, ok := result.(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.Poll.String(), result)
	}

	s.Logger(ctx).Info(fmt.Sprintf("token deployment confirmation result is %s", result))
	s.voter.DeletePoll(ctx, req.Poll)
	s.DeletePendingToken(ctx, req.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeTokenConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.Poll))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &types.VoteConfirmTokenResponse{
			Log: fmt.Sprintf("token %s was discarded", req.Symbol),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	s.nexus.RegisterAsset(ctx, exported.Ethereum.Name, token.Symbol)

	return &types.VoteConfirmTokenResponse{
		Log: fmt.Sprintf("token %s deployment confirmed", token.Symbol)}, nil
}

func (s msgServer) SignDeployToken(c context.Context, req *types.SignDeployTokenRequest) (*types.SignDeployTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chainID := s.GetParams(ctx).Network.Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256([]byte(req.TokenName))[:32])

	data, err := types.CreateDeployTokenCommandData(chainID, commandID, req.TokenName, req.Symbol, req.Decimals, req.Capacity)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := common.Bytes2Hex(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for deploy-token command %s", commandIDHex))
	s.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	s.SetTokenInfo(ctx, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)
	return &types.SignDeployTokenResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignBurnTokens(c context.Context, req *types.SignBurnTokensRequest) (*types.SignBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	deposits := s.GetConfirmedDeposits(ctx)

	if len(deposits) == 0 {
		return &types.SignBurnTokensResponse{}, nil
	}

	chainID := s.GetNetwork(ctx).Params().ChainID

	var burnerInfos []types.BurnerInfo
	seen := map[string]bool{}
	for _, deposit := range deposits {
		if seen[deposit.BurnerAddress.Hex()] {
			continue
		}
		burnerInfo := s.GetBurnerInfo(ctx, common.Address(deposit.BurnerAddress))
		if burnerInfo == nil {
			return nil, fmt.Errorf("no burner info found for address %s", deposit.BurnerAddress.Hex())
		}
		burnerInfos = append(burnerInfos, *burnerInfo)
		seen[deposit.BurnerAddress.Hex()] = true
	}

	data, err := types.CreateBurnCommandData(chainID, ctx.BlockHeight(), burnerInfos)
	if err != nil {
		return nil, err
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for burn command %s", commandIDHex))
	s.SetCommandData(ctx, commandID, data)

	s.Logger(ctx).Info(fmt.Sprintf("signing burn command [%s] for token deposits to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	for _, deposit := range deposits {
		s.DeleteDeposit(ctx, deposit)
		s.SetDeposit(ctx, deposit, types.BURNED)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)
	return &types.SignBurnTokensResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignTx(c context.Context, req *types.SignTxRequest) (*types.SignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	tx := req.UnmarshaledTx()
	txID := tx.Hash().String()
	s.SetUnsignedTx(ctx, txID, tx)
	s.Logger(ctx).Info(fmt.Sprintf("storing raw tx %s", txID))
	hash, err := s.GetHashToSign(ctx, txID)
	if err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("ethereum tx [%s] to sign: %s", txID, hash.Hex()))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, txID),
		),
	)

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, txID, hash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	// if this is the transaction that is deploying Axelar Gateway, calculate and save address
	// TODO: this is something that should be done after the signature has been successfully confirmed
	if tx.To() == nil && bytes.Equal(tx.Data(), s.GetGatewayByteCodes(ctx)) {

		pub, ok := s.signer.GetCurrentKey(ctx, exported.Ethereum, tss.MasterKey)
		if !ok {
			return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
		}

		addr := crypto.CreateAddress(crypto.PubkeyToAddress(pub.Value), tx.Nonce())
		s.SetGatewayAddress(ctx, addr)
	}

	return &types.SignTxResponse{TxID: txID}, nil
}

func (s msgServer) SignPendingTransfers(c context.Context, req *types.SignPendingTransfersRequest) (*types.SignPendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pendingTransfers := s.nexus.GetTransfersForChain(ctx, exported.Ethereum, nexus.Pending)

	if len(pendingTransfers) == 0 {
		return &types.SignPendingTransfersResponse{}, nil
	}

	chainID := s.GetNetwork(ctx).Params().ChainID

	data, err := types.CreateMintCommandData(chainID, pendingTransfers)
	if err != nil {
		return nil, err
	}

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(data)[:32])

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for mint command %s", commandIDHex))
	s.SetCommandData(ctx, commandID, data)

	s.Logger(ctx).Info(fmt.Sprintf("signing mint command [%s] for pending transfers to chain %s", commandIDHex, exported.Ethereum.Name))
	signHash := types.GetEthereumSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	// TODO: Archive pending transfers after signing is completed
	for _, pendingTransfer := range pendingTransfers {
		s.nexus.ArchivePendingTransfer(ctx, pendingTransfer)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &types.SignPendingTransfersResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) SignTransferOwnership(c context.Context, req *types.SignTransferOwnershipRequest) (*types.SignTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chainID := s.GetNetwork(ctx).Params().ChainID

	var commandID types.CommandID
	copy(commandID[:], crypto.Keccak256(req.NewOwner.Bytes())[:32])

	data, err := types.CreateTransferOwnershipCommandData(chainID, commandID, common.Address(req.NewOwner))
	if err != nil {
		return nil, err
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Ethereum, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Ethereum.Name)
	}

	commandIDHex := hex.EncodeToString(commandID[:])
	s.Logger(ctx).Info(fmt.Sprintf("storing data for transfer-ownership command %s", commandIDHex))
	s.SetCommandData(ctx, commandID, data)

	signHash := types.GetEthereumSignHash(data)

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	err = s.signer.StartSign(ctx, s.voter, keyID, commandIDHex, signHash.Bytes(), snapshot)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyCommandID, commandIDHex),
		),
	)

	return &types.SignTransferOwnershipResponse{CommandID: commandID[:]}, nil
}

func (s msgServer) AddChain(c context.Context, req *types.AddChainRequest) (*types.AddChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain := nexus.Chain{
		Name:                  req.Name,
		NativeAsset:           req.NativeAsset,
		SupportsForeignAssets: true,
	}

	if _, found := s.nexus.GetChain(ctx, chain.Name); found {
		return &types.AddChainResponse{}, fmt.Errorf("chain '%s' is already defined", chain.Name)
	}

	s.nexus.SetChain(ctx, chain)
	s.nexus.RegisterAsset(ctx, chain.Name, chain.NativeAsset)

	return &types.AddChainResponse{}, nil
}
