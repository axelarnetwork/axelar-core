package rest

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// query parameters
const (
	QParamFromAddress = "from_address"
	QParamCommandID   = "command_id"
	QParamGasPrice    = "gas_price"
	QParamGasLimit    = "gas_limit"
)

// GetHandlerQueryMasterAddress returns a handler to query the ethereum master address
func GetHandlerQueryMasterAddress(cliCtx client.Context) http.HandlerFunc {
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

// GetHandlerQueryAxelarGatewayAddress returns a handler to query the ethereum gateway contract address
func GetHandlerQueryAxelarGatewayAddress(cliCtx client.Context) http.HandlerFunc {
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

// GetHandlerQueryCommandData returns a handler to query command data by id
func GetHandlerQueryCommandData(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		commandID := mux.Vars(r)[utils.PathVarCommandID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryCommandData, commandID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendCommandTx, commandID).Error())
			return
		}

		var data []byte
		cliCtx.LegacyAmino.MustUnmarshalJSON(res, &data)
		rest.PostProcessResponse(w, cliCtx, common.Bytes2Hex(data))
	}
}

// GetHandlerQueryCreateDeployTx returns a handler to create an ethereum transaction to deploy a smart contract
func GetHandlerQueryCreateDeployTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		gasPrice, ok := parseGasPrice(w, r)
		if !ok {
			return
		}

		gasLimit, ok := parseGasLimit(w, r)
		if !ok {
			return
		}

		params := types.DeployParams{
			GasPrice: gasPrice,
			GasLimit: gasLimit,
		}

		json, err := cliCtx.LegacyAmino.MarshalJSON(params)
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

// GetHandlerQuerySendTx returns a handler to send a transaction to an ethereum wallet to be signed and submitted by a specified account
func GetHandlerQuerySendTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		txID := mux.Vars(r)[utils.PathVarTxID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.SendTx, txID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendTx, txID).Error())
			return
		}

		var result types.SendTxResult
		err = cliCtx.LegacyAmino.UnmarshalJSON(res, &result)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, result)
	}
}

// GetHandlerQuerySendCommandTx returns a handler to send an ethereum transaction containing a smart contract call
// to an ethereum wallet to be signed and submitted by a specified account
func GetHandlerQuerySendCommandTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		fromAddr := r.URL.Query().Get(QParamFromAddress)
		commandIDHex := r.URL.Query().Get(QParamCommandID)

		var commandID types.CommandID
		copy(commandID[:], common.Hex2Bytes(commandIDHex))

		params := types.CommandParams{
			CommandID: commandID,
			Sender:    fromAddr,
		}

		json, err := cliCtx.LegacyAmino.MarshalJSON(params)
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
		cliCtx.LegacyAmino.MustUnmarshalJSON(res, &txHash)
		rest.PostProcessResponse(w, cliCtx, txHash)
	}
}

func parseGasLimit(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	glStr := r.URL.Query().Get(QParamGasLimit)
	gl, err := strconv.ParseUint(glStr, 10, 64)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "could not parse gas limit").Error())
		return 0, false
	}

	return gl, true
}

func parseGasPrice(w http.ResponseWriter, r *http.Request) (sdk.Int, bool) {
	gpStr := r.URL.Query().Get(QParamGasPrice)
	gpBig, ok := big.NewInt(0).SetString(gpStr, 10)
	if !ok {
		rest.WriteErrorResponse(w, http.StatusBadRequest, "could not parse gas price")
		return sdk.Int{}, false
	}

	return sdk.NewIntFromBigInt(gpBig), true
}
