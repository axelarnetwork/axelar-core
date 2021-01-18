package rest

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/btcsuite/btcutil"

	//"github.com/btcsuite/btcutil"
	//sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type CrossChainAddress struct {
	Chain   exported.Chain `json:"chain" yaml:"chain"`
	Address string `json:"address" yaml:"address"`
}

type RespDepositAddress struct {
	Address string `json:"address" yaml:"address"`
}

func QueryDepositAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		chain := balance.ChainFromString(vars["chain"])
		queryData, err := cliCtx.Codec.MarshalJSON(CrossChainAddress{Chain: chain, Address: vars["address"]})
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

func QueryConsolidationAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		out, err := types.OutPointFromStr(extractArgOutpoint(r))
		if err != nil {
			return
		}

		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryConsolidationAddress, out)
		res, _, err := cliCtx.QueryWithData(path, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFConsolidationAddress).Error())
			return
		}

		// the query will return empty if there is no data for this address
		if len(res) == 0 {
			//rest.PostProcessResponse(w, cliCtx, "")
			rest.PostProcessResponse(w, cliCtx, btcutil.AddressPubKey{})
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func QueryTxInfo(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		out, err := types.OutPointFromStr(extractArgOutpoint(r))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		queryData, err := cliCtx.Codec.MarshalJSON(out)
		if err != nil {
			// @TODO wrap failed to marshal outpoint
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryOutInfo), queryData)
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

func QueryRawTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		out, err := types.OutPointFromStr(extractArgOutpoint(r))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		amount, err := denom.ParseSatoshi(r.URL.Query().Get("amount"))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.RawTxParams{
			DepositAddr: r.URL.Query().Get("recipient"),
			OutPoint:    out,
			Satoshi:     amount,
		}

		queryData, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			// @TODO wrap failed to marshal outpoint
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryRawTx), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf(types.ErrFRawTx, out.String()))
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func QuerySendTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		out, err := types.OutPointFromStr(extractArgOutpoint(r))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		queryData, err := cliCtx.Codec.MarshalJSON(out)
		if err != nil {
			// @TODO wrap failed to marshal outpoint
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryRawTx), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf(types.ErrFSendTx, out.String()))
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func extractArgOutpoint(r *http.Request) string {
	return fmt.Sprintf("%s:%s", mux.Vars(r)["txID"], r.URL.Query().Get("voutIdx"))
}