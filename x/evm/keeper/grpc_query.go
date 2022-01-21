package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var _ types.QueryServiceServer = baseKeeper{}

func (k baseKeeper) BurnerInfo(c context.Context, req *types.BurnerInfoRequest) (*types.BurnerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !k.getBaseStore(ctx).Has(subspacePrefix.AppendStr(strings.ToLower(req.Chain))) {
		return nil, sdkerrors.Wrapf(types.ErrEVM, "unkown chain '%s'", req.Chain)
	}

	burnerInfo := k.ForChain(req.Chain).GetBurnerInfo(ctx, common.HexToAddress(req.Address))
	if burnerInfo == nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("unknown burner address '%s'", req.Address))
	}

	return &types.BurnerInfoResponse{BurnerInfo: burnerInfo}, nil
}
