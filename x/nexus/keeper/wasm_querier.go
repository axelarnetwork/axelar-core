package keeper

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// WasmQuerier is a querier for the wasm contracts
type WasmQuerier struct {
	nexus types.Nexus
}

// NewWasmQuerier creates a new WasmQuerier
func NewWasmQuerier(nexus types.Nexus) *WasmQuerier {
	return &WasmQuerier{nexus}
}

// Query handles the wasm queries for the nexus module
func (q WasmQuerier) Query(ctx sdk.Context, req exported.WasmQueryRequest) ([]byte, error) {
	switch {
	case req.TxHashAndNonce != nil:
		txHash, nonce := q.nexus.CurrID(ctx)

		return funcs.Must(json.Marshal(exported.WasmQueryTxHashAndNonceResponse{
			TxHash: txHash,
			Nonce:  nonce,
		})), nil
	case req.IsChainRegistered != nil:
		chainName := exported.ChainName(req.IsChainRegistered.Chain)
		if err := chainName.Validate(); err != nil {
			return nil, err
		}

		_, registered := q.nexus.GetChain(ctx, chainName)
		return funcs.Must(json.Marshal(exported.WasmQueryIsChainRegisteredResponse{
			IsRegistered: registered,
		})), nil

	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Nexus query request"}
	}
}
