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
	QueryParamKeyRole     = "key_role"
	QueryParamKeyID       = "key_id"
)

// GetHandlerQueryLatestBatchedCommands returns a handler to query batched commands by ID
func GetHandlerQueryLatestBatchedCommands(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]

		bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QLatestBatchedCommands, chain))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "could not get the latest batched commands for chain %s", chain).Error())
			return
		}

		var res types.QueryBatchedCommandsResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetHandlerQueryBatchedCommands returns a handler to query batched commands by ID
func GetHandlerQueryBatchedCommands(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		batchedCommandsID := mux.Vars(r)[utils.PathVarBatchedCommandsID]

		bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QBatchedCommands, chain, batchedCommandsID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFBatchedCommands, chain, batchedCommandsID).Error())
			return
		}

		var res types.QueryBatchedCommandsResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetHandlerQueryAddress returns a handler to query an EVM chain address
func GetHandlerQueryAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		keyID := r.URL.Query().Get(QueryParamKeyID)
		keyRole := r.URL.Query().Get(QueryParamKeyRole)

		var query string
		var param string
		switch {
		case keyRole != "" && keyID == "":
			query = keeper.QAddressByKeyRole
			param = keyRole
		case keyRole == "" && keyID != "":
			query = keeper.QAddressByKeyID
			param = keyID
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "one and only one of the two flags key_role and key_id has to be set")
			return
		}

		path := fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, query, chain, param)

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrAddress).Error())
			return
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrAddress).Error())
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
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrAddress).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToAddress(res).Hex())
	}
}

// GetHandlerQueryBytecode returns a handler to fetch the bytecodes of an EVM contract
func GetHandlerQueryBytecode(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]
		contract := mux.Vars(r)[utils.PathVarContract]

		res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QBytecode, chain, contract))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFBytecode, contract).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToHash(res).Hex())
	}
}

// GetHandlerQuerySignedTx returns a handler to fetch an EVM transaction that has been signed by the validators
func GetHandlerQuerySignedTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		chain := mux.Vars(r)[utils.PathVarChain]
		txID := mux.Vars(r)[utils.PathVarTxID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QSignedTx, chain, txID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, types.ErrFSignedTx, txID).Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, common.BytesToHash(res).Hex())
	}
}

// GetHandlerQueryDepositAddress returns a handler to query the state of a deposit address
func GetHandlerQueryDepositAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		recipientChain := mux.Vars(r)[utils.PathVarRecipientChain]
		linkedAddress := mux.Vars(r)[utils.PathVarLinkedAddress]
		symbol := mux.Vars(r)[utils.PathvarSymbol]

		params := types.DepositQueryParams{
			Chain:   recipientChain,
			Address: linkedAddress,
			Symbol:  symbol,
		}
		data := types.ModuleCdc.MustMarshalJSON(&params)

		bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QDepositAddress, chain), data)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDepositState).Error())
			return
		}

		out := common.BytesToAddress(bz)
		rest.PostProcessResponse(w, cliCtx, out.Hex())
	}
}

// GetHandlerQueryDepositState returns a handler to query the state of an ERC20 deposit confirmation
func GetHandlerQueryDepositState(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		txID := common.HexToHash(mux.Vars(r)[utils.PathVarTxID])
		burnerAddress := common.HexToAddress(mux.Vars(r)[utils.PathVarEthereumAddress])
		amount := sdk.NewUintFromString(mux.Vars(r)[utils.PathVarAmount])

		params := types.QueryDepositStateParams{
			TxID:          types.Hash(txID),
			BurnerAddress: types.Address(burnerAddress),
			Amount:        amount.Uint64(),
		}
		data := types.ModuleCdc.MustMarshalJSON(&params)

		bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QDepositState, chain), data)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDepositState).Error())
			return
		}

		var res types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
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
