package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/utils/funcs"
)

type sigHandler struct {
	cdc    codec.Codec
	keeper types.BaseKeeper
}

// NewSigHandler returns the handler for processing signatures delivered by the multisig module
func NewSigHandler(cdc codec.Codec, keeper types.BaseKeeper) multisig.SigHandler {
	return sigHandler{
		cdc:    cdc,
		keeper: keeper,
	}
}

func (s sigHandler) HandleCompleted(ctx sdk.Context, sig codec.ProtoMarshaler, moduleMetadata codec.ProtoMarshaler) error {
	commandBatch, err := s.getCommandBatch(ctx, moduleMetadata)
	if err != nil {
		return err
	}

	funcs.MustNoErr(commandBatch.SetSigned(sig))

	return nil
}

func (s sigHandler) HandleFailed(ctx sdk.Context, moduleMetadata codec.ProtoMarshaler) error {
	commandBatch, err := s.getCommandBatch(ctx, moduleMetadata)
	if err != nil {
		return err
	}

	ok := commandBatch.SetStatus(types.BatchAborted)
	if !ok {
		panic(fmt.Errorf("failed to abort command batch %s", hex.EncodeToString(commandBatch.GetID())))
	}

	return nil
}

func (s sigHandler) getCommandBatch(ctx sdk.Context, moduleMetadata codec.ProtoMarshaler) (types.CommandBatch, error) {
	sigMetadata := moduleMetadata.(*types.SigMetadata)

	if !s.keeper.HasChain(ctx, sigMetadata.Chain) {
		return types.CommandBatch{}, fmt.Errorf("chain %s does not exist as an EVM chain", sigMetadata.Chain)
	}

	ck := s.keeper.ForChain(sigMetadata.Chain)
	commandBatch := ck.GetBatchByID(ctx, sigMetadata.CommandBatchID)
	if !commandBatch.Is(types.BatchSigning) {
		return types.CommandBatch{}, fmt.Errorf("the command batch %s of chain %s is not being signed", hex.EncodeToString(sigMetadata.CommandBatchID), sigMetadata.Chain)
	}

	return commandBatch, nil
}
