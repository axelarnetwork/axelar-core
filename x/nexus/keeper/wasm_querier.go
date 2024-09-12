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
	msgIDGenerator types.MsgIDGenerator
}

// NewWasmQuerier creates a new WasmQuerier
func NewWasmQuerier(msgIDGenerator types.MsgIDGenerator) *WasmQuerier {
	return &WasmQuerier{msgIDGenerator}
}

// Query handles the wasm queries for the nexus module
func (q WasmQuerier) Query(ctx sdk.Context, req exported.WasmQueryRequest) ([]byte, error) {
	if req.TxHashAndNonce != nil {
		txHash, nonce := q.msgIDGenerator.CurrID(ctx)

		return funcs.Must(json.Marshal(exported.WasmQueryTxHashAndNonceResponse{
			TxHash: txHash,
			Nonce:  nonce,
		})), nil
	}

	return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Nexus query request"}
}
