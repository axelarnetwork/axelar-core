package rest

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/utils"
	"net/http"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// RespDepositAddress represents the response of a deposit address query
type RespDepositAddress struct {
	Address string `json:"address" yaml:"address"`
}

const (
	QParamVOutIdx   = "vout_idx"
	QParamBlockHash = "block_hash"
)

// QueryDepositAddress returns a handler to query a deposit address
func QueryDepositAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		queryData, err := cliCtx.Codec.MarshalJSON(types.DepositQueryParams{Chain: vars[utils.PathVarChain], Address: vars[utils.PathVarEthereumAddress]})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryDepositAddress), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDepositAddress).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		resp := RespDepositAddress{Address: string(res)}

		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// QuerySendTransfers returns a handler to send a transaction to Bitcoin
func QuerySendTransfers(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.SendTx), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrFSendTransfers)
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryGetConsolidationTx returns a handler to build a consolidation transaction
func QueryGetConsolidationTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.GetTx), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrFSendTransfers)
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
