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
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// query parameters
const (
	QueryParamFromAddress = "from_address"
	QueryParamCommandID   = "command_id"
	QueryParamGasPrice    = "gas_price"
	QueryParamGasLimit    = "gas_limit"
)

// GetHandlerQueryMasterAddress returns a handler to query an EVM chain master address
func GetHandlerQueryMasterAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QMasterAddress, chain), nil)
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

// GetHandlerQueryNextMasterAddress returns a handler to query an EVM chain next master address
func GetHandlerQueryNextMasterAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QNextMasterAddress, chain), nil)
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

// GetHandlerQueryKeyAddress returns a handler to query to query the EVM address of any key
func GetHandlerQueryKeyAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		keyID := mux.Vars(r)[utils.PathVarKeyID]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryKeyAddress), []byte(keyID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFMasterKey, keyID).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToAddress(res).Hex())
	}
}

// GetHandlerQueryAxelarGatewayAddress returns a handler to query an EVM chain gateway contract address
func GetHandlerQueryAxelarGatewayAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QAxelarGatewayAddress, chain), nil)
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
		chain := mux.Vars(r)[utils.PathVarChain]
		commandID := mux.Vars(r)[utils.PathVarCommandID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QCommandData, chain, commandID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendCommandTx, chain, commandID).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.Bytes2Hex(res))
	}
}

// GetHandlerQueryCreateDeployTx returns a handler to create an EVM chain transaction to deploy a smart contract
func GetHandlerQueryCreateDeployTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		gasPrice, ok := parseGasPrice(w, r)
		if !ok {
			return
		}
		gasLimit, ok := parseGasLimit(w, r)
		if !ok {
			return
		}

		params := types.DeployParams{
			Chain:    chain,
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

// GetHandlerQuerySendTx returns a handler to send a transaction to an EVM chain wallet to be signed and submitted by a specified account
func GetHandlerQuerySendTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]
		txID := mux.Vars(r)[utils.PathVarTxID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.SendTx, chain, txID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendTx, txID).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToHash(res).Hex())
	}
}

// GetHandlerQuerySendCommandTx returns a handler to send an EVM chain transaction containing a smart contract call
// to a wallet to be signed and submitted by a specified account
func GetHandlerQuerySendCommandTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		fromAddr := r.URL.Query().Get(QueryParamFromAddress)
		commandIDHex := r.URL.Query().Get(QueryParamCommandID)

		var commandID types.CommandID
		copy(commandID[:], common.Hex2Bytes(commandIDHex))

		params := types.CommandParams{
			Chain:     chain,
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSendCommandTx, chain, commandIDHex).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToHash(res).Hex())
	}
}

func parseGasLimit(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	glStr := r.URL.Query().Get(QueryParamGasLimit)
	gl, err := strconv.ParseUint(glStr, 10, 64)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "could not parse gas limit").Error())
		return 0, false
	}

	return gl, true
}

func parseGasPrice(w http.ResponseWriter, r *http.Request) (sdk.Int, bool) {
	gpStr := r.URL.Query().Get(QueryParamGasPrice)
	gpBig, ok := big.NewInt(0).SetString(gpStr, 10)
	if !ok {
		rest.WriteErrorResponse(w, http.StatusBadRequest, "could not parse gas price")
		return sdk.Int{}, false
	}

	return sdk.NewIntFromBigInt(gpBig), true
}
