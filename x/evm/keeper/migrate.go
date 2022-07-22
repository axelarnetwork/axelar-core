package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
// - set BurnerCode for external token to nil
// - migrate tss signature to multisig signature
// - set EndBlockerLimit parameter
func GetMigrationHandler(k BaseKeeper, n types.Nexus, s types.Signer, m types.MultisigKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := migrateContractsBytecode(ctx, ck); err != nil {
				return err
			}
		}

		// set external token burner token to nil
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := removeExternalTokenBurnerCode(ctx, ck); err != nil {
				return err
			}

			if err := migrateCommandBatchSignature(ctx, ck, s, m); err != nil {
				return sdkerrors.Wrap(err, fmt.Sprintf("failed to migrate signature for chain %s", chain.Name))
			}

			if err := addEndBlockerLimitParam(ctx, ck); err != nil {
				return err
			}
		}

		return nil
	}
}

func removeExternalTokenBurnerCode(ctx sdk.Context, ck chainKeeper) error {
	for _, meta := range ck.getTokensMetadata(ctx) {
		if !meta.IsExternal {
			continue
		}

		meta.BurnerCode = nil
		ck.setTokenMetadata(ctx, meta)
	}

	return nil
}

func migrateCommandBatchSignature(ctx sdk.Context, ck chainKeeper, signer types.Signer, multisig types.MultisigKeeper) error {
	var commandBatchMetadata types.CommandBatchMetadata
	for commandBatchID := ck.getLatestSignedCommandBatchID(ctx); commandBatchID != nil; commandBatchID = commandBatchMetadata.PrevBatchedCommandsID {
		commandBatchMetadata = ck.getCommandBatchMetadata(ctx, commandBatchID)

		// only migrate secondary key
		tssKey, ok := signer.GetKey(ctx, tss.KeyID(commandBatchMetadata.KeyID))
		if tssKey.Role != tss.SecondaryKey {
			continue
		}

		// only migrate command batch signed by active key
		key, ok := multisig.GetKey(ctx, commandBatchMetadata.KeyID)
		if !ok {
			break
		}

		if commandBatchMetadata.Status != types.BatchSigned {
			if commandBatchMetadata.Status == types.BatchSigning {
				setCommandBatchAborted(ctx, ck, commandBatchMetadata)
			}
			continue
		}

		sigID := hex.EncodeToString(commandBatchMetadata.ID)
		multisigInfo, found := signer.GetMultisigSignInfo(ctx, sigID)
		signInfo, ok := multisigInfo.(*tsstypes.MultisigInfo)
		if !found || !ok {
			return fmt.Errorf("sign info %s not found", sigID)
		}

		payloadHash := commandBatchMetadata.SigHash.Bytes()

		// convert to multisig
		newMultisig := multisigTypes.MultiSig{
			KeyID:       commandBatchMetadata.KeyID,
			PayloadHash: payloadHash,
			Sigs:        make(map[string]multisigTypes.Signature),
		}

		for _, info := range signInfo.Infos {
			if len(info.Data) < 1 {
				return fmt.Errorf("validator %s is participant without signature for signing %s", info.Participant.String(), sigID)
			}

			sigKeyPairs := slices.Map(info.Data, func(bz []byte) tss.SigKeyPair {
				var pair tss.SigKeyPair
				funcs.MustNoErr(pair.Unmarshal(bz))
				return pair
			})

			pubKey, ok := key.GetPubKey(info.Participant)
			if !ok {
				return fmt.Errorf("alidator %s is not a participant for signing %s", info.Participant.String(), sigID)
			}

			signature, err := getSigByPubKey(pubKey, sigKeyPairs)
			if err != nil {
				return err
			}

			if !multisigTypes.Signature(signature).Verify(payloadHash, pubKey) {
				return fmt.Errorf("signature and pubkey mismatch from participant %s for signing %s", info.Participant.String(), sigID)
			}

			newMultisig.Sigs[info.Participant.String()] = signature
		}

		if err := newMultisig.ValidateBasic(); err != nil {
			return err
		}

		setCommandBatchSignature(ctx, ck, commandBatchMetadata, newMultisig)
	}

	return nil
}

func getSigByPubKey(pubKey []byte, sigKeyPairs []tss.SigKeyPair) ([]byte, error) {
	for _, sigKeyPair := range sigKeyPairs {
		if bytes.Equal(sigKeyPair.PubKey, pubKey) {
			return sigKeyPair.Signature, nil
		}
	}

	return nil, fmt.Errorf("failed to find signature for the given pubkey")
}

func setCommandBatchSignature(ctx sdk.Context, ck chainKeeper, commandBatchMetadata types.CommandBatchMetadata, newMultisig multisigTypes.MultiSig) {
	commandBatchMetadata.Signature = funcs.Must(codectypes.NewAnyWithValue(&newMultisig))
	ck.setCommandBatchMetadata(ctx, commandBatchMetadata)
}

func setCommandBatchAborted(ctx sdk.Context, ck chainKeeper, commandBatchMetadata types.CommandBatchMetadata) {
	commandBatchMetadata.Status = types.BatchAborted
	ck.setCommandBatchMetadata(ctx, commandBatchMetadata)
}

func addEndBlockerLimitParam(ctx sdk.Context, ck chainKeeper) error {
	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyEndBlockerLimit, types.DefaultParams()[0].EndBlockerLimit)

	return nil
}

// this function migrates the contracts bytecode to the latest for every existing
// EVM chain. It's crucial whenever contracts are changed between versions and
// DO NOT DELETE
func migrateContractsBytecode(ctx sdk.Context, ck chainKeeper) error {
	bzToken, err := hex.DecodeString(types.Token)
	if err != nil {
		return err
	}

	bzBurnable, err := hex.DecodeString(types.Burnable)
	if err != nil {
		return err
	}

	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyToken, bzToken)
	subspace.Set(ctx, types.KeyBurnable, bzBurnable)

	return nil
}
