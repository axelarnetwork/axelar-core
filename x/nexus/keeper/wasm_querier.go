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
	txIDGenerator types.TxIDGenerator
}

// NewWasmQuerier creates a new WasmQuerier
func NewWasmQuerier(txIDGenerator types.TxIDGenerator) *WasmQuerier {
	return &WasmQuerier{txIDGenerator}
}

// Query handles the wasm queries for the nexus module
func (q WasmQuerier) Query(ctx sdk.Context, req exported.WasmQueryRequest) ([]byte, error) {
	if req.TxID != nil {
		txHash, index := q.txIDGenerator.CurrID(ctx)

		return funcs.Must(json.Marshal(exported.WasmQueryTxIDResponse{
			TxHash: txHash,
			Index:  index,
		})), nil
	}

	return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Nexus query request"}
}
