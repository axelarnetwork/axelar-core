package keeper

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

type WasmQuerier struct {
	txIDGenerator types.TxIDGenerator
}

func NewWasmQuerier(txIDGenerator types.TxIDGenerator) *WasmQuerier {
	return &WasmQuerier{txIDGenerator}
}

func (q WasmQuerier) Query(ctx sdk.Context, req exported.WasmQueryRequest) ([]byte, error) {
	if req.TxID != nil {
		txHash, index := q.txIDGenerator.Curr(ctx)

		return funcs.Must(json.Marshal(exported.WasmQueryTxIDResponse{
			TxHash: txHash,
			Index:  index,
		})), nil
	}

	return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Nexus query request"}
}
