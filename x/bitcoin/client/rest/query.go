package rest

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// query parameters
const (
	QueryParamKeyRole = "key_role"
	QueryParamKeyID   = "key_id"
)

// QueryHandlerDepositAddress returns a handler to query the deposit address for a recipient address on another blockchain
func QueryHandlerDepositAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		params := types.DepositQueryParams{Chain: vars[utils.PathVarChain], Address: vars[utils.PathVarEthereumAddress]}
		path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QDepositAddress)

		bz, _, err := cliCtx.QueryWithData(path, types.ModuleCdc.MustMarshalLengthPrefixed(&params))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrDepositAddr).Error())
			return
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerDepositStatus returns a handler to query the deposit status for a given outpoint
func QueryHandlerDepositStatus(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QDepositStatus, vars[utils.PathVarOutpoint])

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrDepositStatus).Error())
			return
		}

		var res types.QueryDepositStatusResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerConsolidationAddress returns a handler to query the consolidation address
func QueryHandlerConsolidationAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		keyID := r.URL.Query().Get(QueryParamKeyID)
		keyRole := r.URL.Query().Get(QueryParamKeyRole)

		var query string
		var param string
		switch {
		case keyRole != "" && keyID == "":
			query = keeper.QConsolidationAddressByKeyRole
			param = keyRole
		case keyRole == "" && keyID != "":
			query = keeper.QConsolidationAddressByKeyID
			param = keyID
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "one and only one of the two flags key_role and key_id has to be set")
			return
		}

		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, query, param)

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrConsolidationAddr).Error())
			return
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerNextKeyID returns a query handler to get the next assigned key ID
func QueryHandlerNextKeyID(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QNextKeyID, vars[utils.PathVarKeyRole])

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrNextKeyID).Error())
			return
		}

		keyID := string(bz)
		rest.PostProcessResponse(w, cliCtx, keyID)
	}
}

// QueryHandlerMinOutputAmount returns a handler to query the minimum amount allowed for any transaction output in satoshi
func QueryHandlerMinOutputAmount(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QMinOutputAmount)

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrMinOutputAmount).Error())
			return
		}

		minOutputAmount := int64(binary.LittleEndian.Uint64(bz))
		rest.PostProcessResponse(w, cliCtx, strconv.FormatInt(minOutputAmount, 10))
	}
}

// QueryHandlerLatestTx returns a handler to query the latest consolidation transaction of the given tx type
func QueryHandlerLatestTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QLatestTxByTxType, vars[utils.PathVarTxType])

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrLatestTx).Error())
		}

		var res types.QueryTxResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerSignedTx returns a handler to query the signed consolidation transaction of the given transaction hash
func QueryHandlerSignedTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QSignedTx, vars[utils.PathVarTxID])

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrSignedTx).Error())
		}

		var res types.QueryTxResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
