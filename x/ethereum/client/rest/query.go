package rest

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/x/ethereum/client/cli"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func QueryMasterAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryMasterKey), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFMasterKey).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToAddress(res).Hex())
	}
}

func QueryCreateMintTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		contractAddr := mux.Vars(r)["contractAddr"]
		recipient := mux.Vars(r)["recipient"]
		amount, err := cli.ValidMintParams(contractAddr, recipient, r.URL.Query().Get("amount"))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		gasLimit, err := parseGasLimit(w, r)
		if err != nil {
			return
		}

		params := types.MintParams{
			Recipient:    recipient,
			Amount:       amount.Amount,
			ContractAddr: contractAddr,
			GasLimit: gasLimit,
		}

		json, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.CreateMintTx), json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFMintTx).Error())
			return
		}

		err = types.ValidTxJson(res)

		var tx *ethTypes.Transaction
		err = cliCtx.Codec.UnmarshalJSON(res, &tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(cliCtx.Codec.MustMarshalJSON(tx)))
	}
}

func QueryCreateDeployTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bz := common.FromHex(strings.TrimSuffix(string(mux.Vars(r)["byteCode"]), "\n"))

		gasLimit, err := parseGasLimit(w, r)
		if err != nil {
			return
		}

		params := types.DeployParams{
			ByteCode: bz,
			GasLimit: gasLimit,
		}

		json, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.CreateDeployTx), json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDeployTx).Error())
			return
		}

		var result types.DeployResult
		err = cliCtx.Codec.UnmarshalJSON(res, &result)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(cliCtx.Codec.MustMarshalJSON(result.Tx)))
	}
}

func QuerySendTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		txID := mux.Vars(r)["txId"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.SendTx, txID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendTx, txID).Error())
			return
		}

		var out string
		err = cliCtx.Codec.UnmarshalJSON(res, &out)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, out)
	}
}

func QuerySendMintTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		commandID := mux.Vars(r)["commandID"]
		contractAddr := mux.Vars(r)["contractAddr"]
		fromAddr := r.URL.Query().Get("fromAddress")

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s/%s", types.QuerierRoute, keeper.SendMintTx, commandID, fromAddr, contractAddr), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendMintTx, commandID).Error())
			return
		}

		var out string
		err = cliCtx.Codec.UnmarshalJSON(res, &out)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, out)
	}
}

func parseGasLimit(w http.ResponseWriter, r *http.Request) (uint64, error) {
	glStr := r.URL.Query().Get("gasLimit")
	gl, err := strconv.ParseUint(glStr, 10, 64)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "cannot parse ethereum gas limit string: %s", glStr).Error())
		return 0, err
	}

	return gl, nil
}
