package rest

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

func QueryMasterAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryMasterAddress), nil)
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

func QueryAxelarGatewayAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryAxelarGatewayAddress), nil)
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

func QueryCreateDeployTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		gasPrice, err := parseGasPrice(w, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		gasLimit, err := parseGasLimit(w, r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		params := types.DeployParams{
			GasPrice: gasPrice,
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

		txID := mux.Vars(r)["txID"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.SendTx, txID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendTx, txID).Error())
			return
		}

		var result types.SendTxResult
		err = cliCtx.Codec.UnmarshalJSON(res, &result)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, result)
	}
}

const QueryParamContractAddress = "contract_address"
const QueryParamFromAddress = "from_address"
const QueryParamCommandID = "command_id"

func QuerySendCommandTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		contractAddr := mux.Vars(r)[QueryParamContractAddress]
		fromAddr := r.URL.Query().Get(QueryParamFromAddress)
		commandIDHex := r.URL.Query().Get(QueryParamCommandID)

		var commandID types.CommandID
		copy(commandID[:], common.Hex2Bytes(commandIDHex))

		params := types.CommandParams{
			CommandID:    commandID,
			Sender:       fromAddr,
			ContractAddr: contractAddr,
		}

		json, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.SendCommand), json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendCommandTx, commandIDHex).Error())
			return
		}

		var txHash string
		cliCtx.Codec.MustUnmarshalJSON(res, &txHash)
		rest.PostProcessResponse(w, cliCtx, txHash)
	}
}

func parseGasLimit(w http.ResponseWriter, r *http.Request) (uint64, error) {
	glStr := r.URL.Query().Get("gasLimit")
	gl, err := strconv.ParseUint(glStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return gl, nil
}

func parseGasPrice(w http.ResponseWriter, r *http.Request) (sdk.Int, error) {
	gpStr := r.URL.Query().Get("gasPrice")
	gpBig, ok := big.NewInt(0).SetString(gpStr, 10)
	if !ok {
		return sdk.Int{}, fmt.Errorf("could not parse specified gas price")
	}

	return sdk.NewIntFromBigInt(gpBig), nil
}
