package rest

import (
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/wire"

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

// QueryDepositAddress returns a query for a deposit address
func QueryDepositAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		queryData, err := cliCtx.Codec.MarshalJSON(types.DepositQueryParams{Chain: vars["chain"], Address: vars["address"]})
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

// QueryTxInfo returns a query for transaction info
func QueryTxInfo(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		out, err := outPointFromParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		blockHash := r.URL.Query().Get("block_hash")

		queryData, err := cliCtx.Codec.MarshalJSON(out)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryOutInfo, blockHash), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFTxInfo, out.Hash.String(), out.Index).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QuerySendTx returns a query to send a transaction to Bitcoin
func QuerySendTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.SendTx), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrFSendTx)
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func outPointFromParams(r *http.Request) (*wire.OutPoint, error) {
	txId := mux.Vars(r)["txID"]
	idx := r.URL.Query().Get("voutIdx")
	return types.OutPointFromStr(txId + ":" + idx)
}
