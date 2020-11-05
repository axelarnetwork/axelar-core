package keeper

import (
	"fmt"
	"math/big"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) StartSign(ctx sdk.Context, info types.MsgSignStart) error {
	k.Logger(ctx).Info(fmt.Sprintf("TODO not implemented: StartSign: signature [%s] key [%s] ", info.NewSigID, info.KeyID))
	return nil
}

// GetSig returns the signature associated with sigID
// or nil, nil if no such signature exists
// TODO we need a suiable signature struct
// Tendermint uses btcd under the hood:
// https://github.com/tendermint/tendermint/blob/1a8e42d41e9a2a21cb47806a083253ad54c22456/crypto/secp256k1/secp256k1_nocgo.go#L62
// https://github.com/btcsuite/btcd/blob/535f25593d47297f2c7f27fac7725c3b9b05727d/btcec/signature.go#L25-L29
// but we don't want to import btcd everywhere
func (k *Keeper) GetSig(ctx sdk.Context, sigID string) (r *big.Int, s *big.Int) {
	return nil, nil
}
