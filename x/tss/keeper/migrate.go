package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.17 to v0.18. The
// migration includes:
// - migrate sign infos' sigMetadata from JSON to Protobuf
func GetMigrationHandler(k types.TSSKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		if err := migrateSignInfo(ctx, k); err != nil {
			return err
		}

		return nil
	}
}

func migrateSignInfo(ctx sdk.Context, k types.TSSKeeper) error {
	iter := k.(Keeper).getStore(ctx).Iterator(infoForSigPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var signInfo exported.SignInfo
		iter.UnmarshalValue(&signInfo)

		var sigMetadata evmTypes.SigMetadata
		if err := types.ModuleCdc.UnmarshalJSON([]byte(signInfo.Metadata), &sigMetadata); err != nil {
			continue
		}

		sigMetadataProto, err := codectypes.NewAnyWithValue(&sigMetadata)
		if err != nil {
			return err
		}

		signInfo.ModuleMetadata = sigMetadataProto
		signInfo.Metadata = ""

		k.SetInfoForSig(ctx, signInfo.SigID, signInfo)
	}

	return nil
}
