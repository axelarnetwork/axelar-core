package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
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

func queryChains(ctx sdk.Context, n types.Nexus) []string {
	chains := []string{}
	for _, c := range n.GetChains(ctx) {
		if c.Module == types.ModuleName {
			chains = append(chains, c.Name)
		}
	}

	return chains
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

	if !q.keeper.HasChain(ctx, req.Chain) {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("%s is not a registered chain", req.Chain)).Error())
	}
	ck := q.keeper.ForChain(req.Chain)

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

	_, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, "unknown chain")

	}

	ck := q.keeper.ForChain(string(req.Chain))
	height, ok := ck.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, "could not get confirmation height")
	}

	return &types.ConfirmationHeightResponse{Height: height}, nil
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

	_, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("chain %s not found", req.Chain))
	}

	ck := q.keeper.ForChain(req.Chain)

	var commands []types.QueryCommandResponse
	for _, cmd := range ck.GetPendingCommands(ctx) {
		cmdResp, err := GetCommandResponse(ctx, ck.GetName(), q.nexus, cmd)
		if err != nil {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		commands = append(commands, cmdResp)
	}

	return &types.PendingCommandsResponse{Commands: commands}, nil
}
