package rest

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/axelarnetwork/axelar-core/utils"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/types/rest"
)

// query parameters
const (
	QueryParamFeeRate = "fee_rate"
)

// QueryHandlerDepositAddress returns a handler to query a deposit address
func QueryHandlerDepositAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		queryData, err := cliCtx.LegacyAmino.MarshalJSON(types.DepositQueryParams{Chain: vars[utils.PathVarChain], Address: vars[utils.PathVarEthereumAddress]})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QDepositAddress), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFMasterKey).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(res))
	}
}

// QueryHandlerMinimumWithdrawAmount returns a handler to query the minimum amount to withdraw in satoshi
func QueryHandlerMinimumWithdrawAmount(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QMinimumWithdrawAmount), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFMasterKey).Error())
			return
		}

		response := int64(binary.LittleEndian.Uint64(res))
		rest.PostProcessResponse(w, cliCtx, strconv.FormatInt(response, 10))
	}
}

// QueryHandlerMasterAddress returns a handler to query the segwit address of the master key
func QueryHandlerMasterAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QMasterAddress), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var resp types.QueryMasterAddressResponse
		err = resp.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// QueryHandlerKeyConsolidationAddress  returns a handler to query the consolidation segwit address of any key
func QueryHandlerKeyConsolidationAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		keyID := mux.Vars(r)[utils.PathVarKeyID]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QKeyConsolidationAddress), []byte(keyID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(res))
	}
}

// GetConsolidationTxResult models the QueryRawTxResponse from keeper.GetConsolidationTx as a JSON response
type GetConsolidationTxResult struct {
	State types.SignState `json:"state"`
	RawTx string          `json:"raw_tx"`
}

// QueryHandlerGetConsolidationTx returns a handler to build a consolidation transaction
func QueryHandlerGetConsolidationTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QConsolidationTx), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrFGetRawTx)
			return
		}

		var proto types.QueryRawTxResponse
		err = proto.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to unmarshal QueryRawTxResponse: "+err.Error())
			return
		}

		result := GetConsolidationTxResult{
			State: proto.GetState(),
			RawTx: proto.GetRawTx(),
		}

		rest.PostProcessResponse(w, cliCtx, result)
	}
}

// QueryHandlerGetPayForConsolidationTx returns a handler to build a transaction that pays for the consolidation transaction
func QueryHandlerGetPayForConsolidationTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// Parse fee rate
		feeRateStr := r.URL.Query().Get(QueryParamFeeRate)
		if feeRateStr == "" {
			feeRateStr = "0" // fee is optional and defaults to zero
		}

		feeRate, err := strconv.ParseInt(feeRateStr, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrFInvalidFeeRate)
			return
		}

		bz := make([]byte, 8)
		binary.LittleEndian.PutUint64(bz, uint64(feeRate))

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QPayForConsolidationTx), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFGetPayForRawTx).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(res))
	}
}
