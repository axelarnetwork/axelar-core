package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"

	evmclient "github.com/axelarnetwork/axelar-core/x/evm/client"
)

// query parameters
const (
	QueryParamKeyRole = "key_role"
	QueryParamKeyID   = "key_id"
	QueryParamSymbol  = keeper.BySymbol
	QueryParamAsset   = keeper.ByAsset
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
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetHandlerQueryCommand returns a handler to get the command with the given ID on the specified chain
func GetHandlerQueryCommand(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		id := mux.Vars(r)[utils.PathVarCommandID]

		res, err := evmclient.QueryCommand(cliCtx, chain, id)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

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
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
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
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetHandlerQueryTokenAddress returns a handler to query an EVM chain address
func GetHandlerQueryTokenAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]

		symbol := r.URL.Query().Get(QueryParamSymbol)
		asset := r.URL.Query().Get(QueryParamAsset)

		var bz []byte
		var err error
		switch {
		case symbol != "" && asset == "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QTokenAddressBySymbol, chain, symbol))
		case symbol == "" && asset != "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QTokenAddressByAsset, chain, asset))
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "lookup must be either by asset name or symbol")
			return
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFTokenAddress).Error())
			return
		}

		var res types.QueryTokenAddressResponse
		types.ModuleCdc.UnmarshalLengthPrefixed(bz, &res)

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

		bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QNextMasterAddress, chain))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrAddress).Error())
			return
		}

		var res types.QueryAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
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

		rest.PostProcessResponse(w, cliCtx, "0x"+common.Bytes2Hex(res))
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
			Amount:        amount.String(),
		}
		data := types.ModuleCdc.MustMarshalJSON(&params)

		bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QDepositState, chain), data)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDepositState).Error())
			return
		}

		var res types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}



