package keeper

import (
	"bytes"
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
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc querier
type Querier struct {
	keeper   types.BaseKeeper
	nexus    types.Nexus
	signer   types.Signer
	multisig types.MultisigKeeper
}

// NewGRPCQuerier returns a new Querier
func NewGRPCQuerier(k types.BaseKeeper, n types.Nexus, s types.Signer, multisig types.MultisigKeeper) Querier {
	return Querier{
		keeper:   k,
		nexus:    n,
		signer:   s,
		multisig: multisig,
	}
}

func queryChains(ctx sdk.Context, n types.Nexus) []nexustypes.ChainName {
	chains := slices.Filter(n.GetChains(ctx), types.IsEVMChain)

	return slices.Map(chains, nexustypes.Chain.GetName)
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

// optimizeSignatureSet returns optimized signature set, sorted in ascending order by corresponding evm address
func optimizeSignatureSet(operators []types.Operator, minPassingWeight sdk.Uint) [][]byte {
	sort.SliceStable(operators, func(i, j int) bool {
		return operators[i].Weight.GT(operators[j].Weight)
	})

	cumWeight := sdk.ZeroUint()
	operators = slices.Filter(operators, func(operator types.Operator) bool {
		if cumWeight.GTE(minPassingWeight) {
			return false
		}

		cumWeight = cumWeight.Add(operator.Weight)
		return true
	})

	sort.SliceStable(operators, func(i, j int) bool {
		return bytes.Compare(operators[i].Address.Bytes(), operators[j].Address.Bytes()) < 0
	})

	return slices.Map(operators, func(operator types.Operator) []byte { return operator.Signature })
}

func getProof(key multisig.Key, signature multisig.MultiSig) ([]common.Address, []sdk.Uint, sdk.Uint, [][]byte) {
	participantsWithSigs := slices.Filter(key.GetParticipants(), func(v sdk.ValAddress) bool {
		_, ok := signature.GetSignature(v)
		return ok
	})

	operators := slices.Map(participantsWithSigs, func(val sdk.ValAddress) types.Operator {
		pubKey := funcs.MustOk(key.GetPubKey(val)).ToECDSAPubKey()
		evmSignature := funcs.Must(types.ToSignature(funcs.MustOk(signature.GetSignature(val)), common.BytesToHash(signature.GetPayloadHash()), pubKey))

		return types.Operator{
			Address:   crypto.PubkeyToAddress(pubKey),
			Signature: evmSignature.ToHomesteadSig(),
			Weight:    key.GetWeight(val),
		}
	})

	addresses, weights, threshold := types.GetMultisigAddressesAndWeights(key)
	signatures := optimizeSignatureSet(operators, key.GetMinPassingWeight())

	return addresses, weights, threshold, signatures
}

func getExecuteDataAndSigs(ctx sdk.Context, multisigK types.MultisigKeeper, commandBatch types.CommandBatch, signature multisig.MultiSig) ([]byte, types.Proof, error) {
	key := funcs.MustOk(multisigK.GetKey(ctx, signature.GetKeyID()))

	addresses, weights, threshold, signatures := getProof(key, signature)

	executeData, err := types.CreateExecuteDataMultisig(commandBatch.GetData(), addresses, weights, threshold, signatures)
	if err != nil {
		return nil, types.Proof{}, fmt.Errorf("could not create transaction data: %s", err)
	}

	proof := types.Proof{
		Addresses:  slices.Map(addresses, common.Address.Hex),
		Weights:    slices.Map(weights, sdk.Uint.String),
		Threshold:  threshold.String(),
		Signatures: slices.Map(signatures, hex.EncodeToString),
	}

	return executeData, proof, nil
}

func commandBatchToResp(ctx sdk.Context, commandBatch types.CommandBatch, multisigK types.MultisigKeeper) (types.BatchedCommandsResponse, error) {
	id := hex.EncodeToString(commandBatch.GetID())

	prevID := ""
	if commandBatch.GetPrevBatchedCommandsID() != nil {
		prevID = hex.EncodeToString(commandBatch.GetPrevBatchedCommandsID())
	}

	commandIDs := slices.Map(commandBatch.GetCommandIDs(), types.CommandID.Hex)

	switch {
	case commandBatch.Is(types.BatchSigned) && commandBatch.GetSignature() != nil: // check signature for unmigrated batches
		signature := commandBatch.GetSignature().(multisig.MultiSig)
		executeData, proof, err := getExecuteDataAndSigs(ctx, multisigK, commandBatch, signature)
		if err != nil {
			return types.BatchedCommandsResponse{}, sdkerrors.Wrap(types.ErrEVM, err.Error())
		}

		return types.BatchedCommandsResponse{
			ID:                    id,
			Data:                  hex.EncodeToString(commandBatch.GetData()),
			Status:                commandBatch.GetStatus(),
			KeyID:                 commandBatch.GetKeyID(),
			ExecuteData:           hex.EncodeToString(executeData),
			PrevBatchedCommandsID: prevID,
			CommandIDs:            commandIDs,
			Proof:                 &proof,
		}, nil
	default:
		return types.BatchedCommandsResponse{
			ID:                    id,
			Data:                  hex.EncodeToString(commandBatch.GetData()),
			Status:                commandBatch.GetStatus(),
			KeyID:                 commandBatch.GetKeyID(),
			ExecuteData:           "",
			PrevBatchedCommandsID: prevID,
			CommandIDs:            commandIDs,
			Proof:                 nil,
		}, nil
	}
}

func getCommandBatchSig(pair tss.SigKeyPair, batchedCommands types.Hash) (types.Signature, error) {
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

	var commandBatch types.CommandBatch
	switch req.Id {
	case "":
		commandBatch = ck.GetLatestCommandBatch(ctx)
		if commandBatch.Is(types.BatchNonExistent) {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not get the latest batched commands for chain %s", req.Chain)).Error())
		}
	default:
		commandBatchID, err := hex.DecodeString(req.Id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("invalid batched commands ID: %v", err)).Error())
		}

		commandBatch = ck.GetBatchByID(ctx, commandBatchID)
		if commandBatch.Is(types.BatchNonExistent) {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("batched commands with ID %s not found", req.Id)).Error())
		}
	}

	resp, err := commandBatchToResp(ctx, commandBatch, q.multisig)
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
	if _, ok := n.GetChain(ctx, nexustypes.ChainName(k.GetName())); !ok {
		return -1, fmt.Sprintf("%s is not a registered chain", k.GetName()), codes.NotFound
	}

	_, state, ok := k.GetDeposit(ctx, params.TxID, params.BurnerAddress)
	if !ok {
		return types.DepositStatus_None, "deposit transaction is not confirmed", codes.OK
	}

	switch state {
	case types.DepositStatus_Confirmed:
		return types.DepositStatus_Confirmed, "deposit transaction is confirmed", codes.OK
	case types.DepositStatus_Burned:
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

func queryAddressByKeyID(ctx sdk.Context, multisig types.MultisigKeeper, chain nexustypes.Chain, keyID multisig.KeyID) (types.KeyAddressResponse, error) {
	key, ok := multisig.GetKey(ctx, keyID)
	if !ok {
		return types.KeyAddressResponse{}, sdkerrors.Wrapf(types.ErrEVM, "key %s not found for chain %s", keyID, chain.Name)
	}

	addresses, weights, threshold := types.GetMultisigAddressesAndWeights(key)
	weightedAddresses := make([]types.KeyAddressResponse_WeightedAddress, 0, len(weights))
	for i, address := range addresses {
		weightedAddresses = append(weightedAddresses, types.KeyAddressResponse_WeightedAddress{
			Address: address.Hex(),
			Weight:  weights[i].String(),
		})
	}

	return types.KeyAddressResponse{
		KeyID:     keyID,
		Addresses: weightedAddresses,
		Threshold: threshold.String(),
	}, nil
}

// KeyAddress returns the address the specified key for the specified chain
func (q Querier) KeyAddress(c context.Context, req *types.KeyAddressRequest) (*types.KeyAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, nexustypes.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}

	keyID := req.KeyID
	if keyID == "" {
		keyID, ok = q.multisig.GetCurrentKeyID(ctx, chain.Name)
		if !ok {
			return nil, status.Error(codes.NotFound, sdkerrors.Wrapf(types.ErrEVM, "current key not found for chain %s", req.Chain).Error())
		}
	}

	res, err := queryAddressByKeyID(ctx, q.multisig, chain, keyID)
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

	res := types.ERC20TokensResponse{
		Tokens: slices.Map(tokens, func(token types.ERC20Token) types.ERC20TokensResponse_Token {
			return types.ERC20TokensResponse_Token{
				Asset:  token.GetAsset(),
				Symbol: token.GetDetails().Symbol,
			}
		}),
	}

	return &res, nil
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

	var burnerCodeHashHex string
	if burnerCodeHash, ok := token.GetBurnerCodeHash(); ok {
		burnerCodeHashHex = burnerCodeHash.Hex()
	}

	return &types.TokenInfoResponse{
		Asset:          token.GetAsset(),
		Details:        token.GetDetails(),
		Address:        token.GetAddress().Hex(),
		Confirmed:      token.Is(types.Confirmed),
		IsExternal:     token.IsExternal(),
		BurnerCodeHash: burnerCodeHashHex,
	}, nil
}
