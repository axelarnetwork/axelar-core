package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.QueryServer = Keeper{}

// GovernanceKey returns the multisig governance key
func (k Keeper) GovernanceKey(c context.Context, req *types.QueryGovernanceKeyRequest) (*types.QueryGovernanceKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	governanceKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		return nil, fmt.Errorf("governance key not set")
	}

	var pubKeys []string
	for _, pk := range governanceKey.GetPubKeys() {
		bz, _ := types.ModuleCdc.MarshalInterfaceJSON(pk)
		pubKeys = append(pubKeys, string(bz))
	}

	return &types.QueryGovernanceKeyResponse{
		GovernanceKey: governanceKey,
	}, nil
}
