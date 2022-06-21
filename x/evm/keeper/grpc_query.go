package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc querier
type Querier struct {
	keeper types.BaseKeeper
	nexus  types.Nexus
	signer types.Signer
}

// NewGRPCQuerier returns a new Querier
func NewGRPCQuerier(k types.BaseKeeper, n types.Nexus, s types.Signer) Querier {
	return Querier{
		keeper: k,
		nexus:  n,
		signer: s,
	}
}

func queryChains(ctx sdk.Context, n types.Nexus) []nexustypes.ChainName {
	chains := slices.Filter(n.GetChains(ctx), types.IsEVMChain)

	return slices.Map(chains, func(c nexustypes.Chain) nexustypes.ChainName {
		return c.Name
	})
}

// Chains returns the available evm chains
func (q Querier) Chains(c context.Context, req *types.ChainsRequest) (*types.ChainsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chains := queryChains(ctx, q.nexus)

	return &types.ChainsResponse{Chains: chains}, nil
}

// BurnerInfo implements the burner info grpc query
func (q Querier) BurnerInfo(c context.Context, req *types.BurnerInfoRequest) (*types.BurnerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chains := queryChains(ctx, q.nexus)

	for _, chain := range chains {
		ck := q.keeper.ForChain(chain)
		burnerInfo := ck.GetBurnerInfo(ctx, req.Address)
		if burnerInfo != nil {
			return &types.BurnerInfoResponse{Chain: ck.GetParams(ctx).Chain, BurnerInfo: burnerInfo}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "unknown address")
}

func batchedCommandsToQueryResp(ctx sdk.Context, batchedCommands types.CommandBatch, s types.Signer) (types.BatchedCommandsResponse, error) {
	batchedCommandsIDHex := hex.EncodeToString(batchedCommands.GetID())
	prevBatchedCommandsIDHex := ""
	if batchedCommands.GetPrevBatchedCommandsID() != nil {
		prevBatchedCommandsIDHex = hex.EncodeToString(batchedCommands.GetPrevBatchedCommandsID())
	}

	var commandIDs []string
	for _, id := range batchedCommands.GetCommandIDs() {
		commandIDs = append(commandIDs, id.Hex())
	}

	var resp types.BatchedCommandsResponse

	switch {
	case batchedCommands.Is(types.BatchSigned):
		sig, sigStatus := s.GetSig(ctx, batchedCommandsIDHex)
		if sigStatus != tss.SigStatus_Signed {
			return resp, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not find a corresponding signature for sig ID %s", batchedCommandsIDHex))
		}

		var executeData []byte
		var signatures []string
		switch signature := sig.GetSig().(type) {
		case *tss.Signature_SingleSig_:
			batchedCommandsSig, err := getBatchedCommandsSig(signature.SingleSig.SigKeyPair, batchedCommands.GetSigHash())
			if err != nil {
				return resp, err
			}

			executeData, err = types.CreateExecuteDataSinglesig(batchedCommands.GetData(), batchedCommandsSig)
			if err != nil {
				return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
			}

			signatures = append(signatures, hex.EncodeToString(batchedCommandsSig[:]))
		case *tss.Signature_MultiSig_:
			var batchedCmdSigs []types.Signature
			var err error

			sigKeyPairs := types.SigKeyPairs(signature.MultiSig.SigKeyPairs)
			sort.Stable(sigKeyPairs)

			for _, pair := range sigKeyPairs {
				batchedCommandsSig, err := getBatchedCommandsSig(pair, batchedCommands.GetSigHash())
				if err != nil {
					return resp, err
				}

				batchedCmdSigs = append(batchedCmdSigs, batchedCommandsSig)
				signatures = append(signatures, hex.EncodeToString(batchedCommandsSig[:]))
			}

			executeData, err = types.CreateExecuteDataMultisig(batchedCommands.GetData(), batchedCmdSigs...)
			if err != nil {
				return resp, sdkerrors.Wrapf(types.ErrEVM, "could not create transaction data: %s", err)
			}
		}

		resp = types.BatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             signatures,
			ExecuteData:           hex.EncodeToString(executeData),
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
			CommandIDs:            commandIDs,
		}
	default:
		resp = types.BatchedCommandsResponse{
			ID:                    batchedCommandsIDHex,
			Data:                  hex.EncodeToString(batchedCommands.GetData()),
			Status:                batchedCommands.GetStatus(),
			KeyID:                 batchedCommands.GetKeyID(),
			Signature:             nil,
			ExecuteData:           "",
			PrevBatchedCommandsID: prevBatchedCommandsIDHex,
			CommandIDs:            commandIDs,
		}
	}

	return resp, nil
}

func getBatchedCommandsSig(pair tss.SigKeyPair, batchedCommands types.Hash) (types.Signature, error) {
	pk, err := pair.GetKey()
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not parse pub key: %v", err))
	}

	sig, err := pair.GetSig()
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not parse signature: %v", err))
	}

	batchedCommandsSig, err := types.ToSignature(sig, common.Hash(batchedCommands), pk)
	if err != nil {
		return types.Signature{}, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}
	return batchedCommandsSig, nil
}

// BatchedCommands implements the batched commands query
// If BatchedCommandsResponse.Id is set, it returns the latest batched commands with the specified id.
// Otherwise returns the latest batched commands.
func (q Querier) BatchedCommands(c context.Context, req *types.BatchedCommandsRequest) (*types.BatchedCommandsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !q.keeper.HasChain(ctx, nexustypes.ChainName(req.Chain)) {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}
	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))

	var batchedCommands types.CommandBatch
	if req.Id == "" {
		batchedCommands = ck.GetLatestCommandBatch(ctx)
		if batchedCommands.Is(types.BatchNonExistent) {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not get the latest batched commands for chain %s", req.Chain)).Error())
		}
	} else {
		batchedCommandsID, err := hex.DecodeString(req.Id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("invalid batched commands ID: %v", err)).Error())
		}

		batchedCommands = ck.GetBatchByID(ctx, batchedCommandsID)
		if batchedCommands.Is(types.BatchNonExistent) {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("batched commands with ID %s not found", req.Id)).Error())
		}
	}

	resp, err := batchedCommandsToQueryResp(ctx, batchedCommands, q.signer)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &resp, nil
}

// ConfirmationHeight implements the confirmation height grpc query
func (q Querier) ConfirmationHeight(c context.Context, req *types.ConfirmationHeightRequest) (*types.ConfirmationHeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, "unknown chain")

	}

	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))
	height, ok := ck.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, "could not get confirmation height")
	}

	return &types.ConfirmationHeightResponse{Height: height}, nil
}

// Event implements the query for an event at a chain based on the event's ID
func (q Querier) Event(c context.Context, req *types.EventRequest) (*types.EventResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if !q.keeper.HasChain(ctx, nexustypes.ChainName(req.Chain)) {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("[%s] is not a registered chain", req.Chain)).Error())
	}

	event, ok := q.keeper.ForChain(nexustypes.ChainName(req.Chain)).GetEvent(ctx, types.EventID(req.EventId))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("no event with ID [%s] was found", req.EventId)).Error())
	}

	return &types.EventResponse{Event: &event}, nil
}

func queryDepositState(ctx sdk.Context, k types.ChainKeeper, n types.Nexus, params *types.QueryDepositStateParams) (types.DepositStatus, string, codes.Code) {
	_, ok := n.GetChain(ctx, nexustypes.ChainName(k.GetName()))
	if !ok {
		return -1, fmt.Sprintf("%s is not a registered chain", k.GetName()), codes.NotFound
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%s", params.TxID.Hex(), params.BurnerAddress.Hex(), params.Amount))
	_, isPending := k.GetPendingDeposit(ctx, pollKey)
	_, state, ok := k.GetDeposit(ctx, common.Hash(params.TxID), common.Address(params.BurnerAddress))

	switch {
	case isPending:
		return types.DepositStatus_Pending, "deposit transaction is waiting for confirmation", codes.OK
	case !isPending && !ok:
		return types.DepositStatus_None, "deposit transaction is not confirmed", codes.OK
	case state == types.DepositStatus_Confirmed:
		return types.DepositStatus_Confirmed, "deposit transaction is confirmed", codes.OK
	case state == types.DepositStatus_Burned:
		return types.DepositStatus_Burned, "deposit has been transferred to the destination chain", codes.OK
	default:
		return -1, "deposit is in an unexpected state", codes.Internal
	}
}

// DepositState fetches the state of a deposit confirmation using a grpc query
func (q Querier) DepositState(c context.Context, req *types.DepositStateRequest) (*types.DepositStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ck := q.keeper.ForChain(req.Chain)

	s, log, code := queryDepositState(ctx, ck, q.nexus, req.Params)
	if code != codes.OK {
		return nil, status.Error(code, log)
	}

	return &types.DepositStateResponse{Status: s}, nil
}

// PendingCommands returns the pending commands from a gateway
func (q Querier) PendingCommands(c context.Context, req *types.PendingCommandsRequest) (*types.PendingCommandsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("chain %s not found", req.Chain))
	}

	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))

	var commands []types.QueryCommandResponse
	for _, cmd := range ck.GetPendingCommands(ctx) {
		cmdResp, err := GetCommandResponse(cmd)
		if err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		commands = append(commands, cmdResp)
	}

	return &types.PendingCommandsResponse{Commands: commands}, nil
}

func queryAddressByKeyID(ctx sdk.Context, s types.Signer, chain nexustypes.Chain, keyID tss.KeyID) (types.KeyAddressResponse, error) {
	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return types.KeyAddressResponse{}, sdkerrors.Wrapf(types.ErrEVM, "threshold key %s not found", keyID)
	}

	switch chain.KeyType {
	case tss.Multisig:
		multisigPubKey, err := key.GetMultisigPubKey()
		if err != nil {
			return types.KeyAddressResponse{}, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		addressStrs := make([]string, len(multisigPubKey))
		for i, address := range types.KeysToAddresses(multisigPubKey...) {
			addressStrs[i] = address.Hex()
		}

		threshold := uint32(key.GetMultisigKey().Threshold)

		resp := types.KeyAddressResponse{
			Address: &types.KeyAddressResponse_MultisigAddresses_{MultisigAddresses: &types.KeyAddressResponse_MultisigAddresses{Addresses: addressStrs, Threshold: threshold}},
			KeyID:   keyID,
		}

		return resp, nil
	case tss.Threshold:
		pk, err := key.GetECDSAPubKey()
		if err != nil {
			return types.KeyAddressResponse{}, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		address := crypto.PubkeyToAddress(pk)
		resp := types.KeyAddressResponse{
			Address: &types.KeyAddressResponse_ThresholdAddress_{ThresholdAddress: &types.KeyAddressResponse_ThresholdAddress{Address: address.Hex()}},
			KeyID:   key.ID,
		}

		return resp, nil
	default:
		return types.KeyAddressResponse{}, sdkerrors.Wrapf(types.ErrEVM, "unknown key type %s of chain %s", chain.KeyType, chain.Name)
	}
}

// KeyAddress returns the address the specified key for the specified chain
func (q Querier) KeyAddress(c context.Context, req *types.KeyAddressRequest) (*types.KeyAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}

	var keyID tss.KeyID
	switch key := req.Key.(type) {
	case *types.KeyAddressRequest_KeyID:
		keyID = key.KeyID
	case *types.KeyAddressRequest_Role:
		keyID, ok = q.signer.GetCurrentKeyID(ctx, chain, keyRole)
		if !ok {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrapf(types.ErrEVM, "key not found for chain %s", req.Chain).Error())
		}
	}

	res, err := queryAddressByKeyID(ctx, q.signer, chain, keyID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &res, nil
}

// GatewayAddress returns the axelar gateway address for the specified chain
func (q Querier) GatewayAddress(c context.Context, req *types.GatewayAddressRequest) (*types.GatewayAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !q.keeper.HasChain(ctx, nexustypes.ChainName(req.Chain)) {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}

	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))

	address, ok := ck.GetGatewayAddress(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("axelar gateway not set for chain [%s]", req.Chain)).Error())
	}

	return &types.GatewayAddressResponse{Address: address.Hex()}, nil
}

// Bytecode returns the bytecode of a specified contract and chain
func (q Querier) Bytecode(c context.Context, req *types.BytecodeRequest) (*types.BytecodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain)); !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}

	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))

	var bytecode []byte
	switch strings.ToLower(req.Contract) {
	case BCToken:
		bytecode, _ = ck.GetTokenByteCode(ctx)
	case BCBurner:
		bytecode, _ = ck.GetBurnerByteCode(ctx)
	default:
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not retrieve bytecode for chain %s", req.Chain)).Error())
	}

	return &types.BytecodeResponse{Bytecode: fmt.Sprintf("0x" + common.Bytes2Hex(bytecode))}, nil
}

// ERC20Tokens returns the ERC20 tokens registered for a chain
func (q Querier) ERC20Tokens(c context.Context, req *types.ERC20TokensRequest) (*types.ERC20TokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, fmt.Errorf("chain %s not found", req.Chain)
	}

	if !types.IsEVMChain(chain) {
		return nil, fmt.Errorf("%s not an EVM chain", chain.Name)
	}

	ck := q.keeper.ForChain(chain.Name)

	tokens := ck.GetTokens(ctx)
	switch req.Type {
	case types.External:
		tokens = slices.Filter(tokens, types.ERC20Token.IsExternal)
	case types.Internal:
		tokens = slices.Filter(tokens, func(token types.ERC20Token) bool { return !token.IsExternal() })
	default:
		// no filtering when retrieving all tokens
	}

	assets := slices.Map(tokens, types.ERC20Token.GetAsset)

	return &types.ERC20TokensResponse{Assets: assets}, nil
}

// TokenInfo returns the token info for a registered asset
func (q Querier) TokenInfo(c context.Context, req *types.TokenInfoRequest) (*types.TokenInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, fmt.Errorf("chain %s not found", req.Chain)
	}

	if !types.IsEVMChain(chain) {
		return nil, fmt.Errorf("%s is not an EVM chain", chain.Name)
	}

	ck := q.keeper.ForChain(nexustypes.ChainName(req.Chain))

	var token types.ERC20Token
	switch findBy := req.GetFindBy().(type) {
	case *types.TokenInfoRequest_Asset:
		token = ck.GetERC20TokenByAsset(ctx, findBy.Asset)
		if token.Is(types.NonExistent) {
			return nil, fmt.Errorf("%s is not a registered asset for chain %s", req.GetAsset(), chain.Name)
		}
	case *types.TokenInfoRequest_Symbol:
		token = ck.GetERC20TokenBySymbol(ctx, findBy.Symbol)
		if token.Is(types.NonExistent) {
			return nil, fmt.Errorf("%s is not a registered symbol for chain %s", req.GetSymbol(), chain.Name)
		}
	}

	return &types.TokenInfoResponse{
		Asset:      token.GetAsset(),
		Details:    token.GetDetails(),
		Address:    token.GetAddress().Hex(),
		Confirmed:  token.Is(types.Confirmed),
		IsExternal: token.IsExternal(),
	}, nil
}
