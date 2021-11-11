package utils

import (
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

// routes
const (
	PathVarChain             = "Chain"
	PathVarRecipientChain    = "RecipientChain"
	PathVarContract          = "Contract"
	PathVarCosmosAddress     = "CosmosAddress"
	PathVarCounter           = "Counter"
	PathVarAmount            = "Amount"
	PathVarLinkedAddress     = "LinkedAddress"
	PathVarEthereumAddress   = "EthereumAddress"
	PathVarTxID              = "TxID"
	PathVarCommandID         = "CommandID"
	PathVarBatchedCommandsID = "BatchedCommandsID"
	PathVarKeyRole           = "KeyRole"
	PathVarTxType            = "txType"
	PathVarKeyID             = "KeyID"
	PathVarSigID             = "SigID"
	PathVarOutpoint          = "Outpoint"
	PathvarSymbol            = "Symbol"
	PathVarAsset             = "Asset"
)

// ExtractReqSender extracts the sender address from an SDK base request
func ExtractReqSender(w http.ResponseWriter, req rest.BaseReq) (sdk.AccAddress, bool) {
	sender, err := sdk.AccAddressFromBech32(req.From)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return sender, true
}

// RegisterTxHandlerFn returns a function to register rest routes with the given router
func RegisterTxHandlerFn(r *mux.Router, moduleRoute string) func(http.HandlerFunc, string, ...string) {
	return func(handler http.HandlerFunc, method string, pathVars ...string) {
		path := appendPathVars(fmt.Sprintf("/tx/%s/%s", moduleRoute, method), pathVars)
		r.HandleFunc(path, handler).Methods("POST")
	}
}

// RegisterQueryHandlerFn returns a function to register query routes with the given router
func RegisterQueryHandlerFn(r *mux.Router, moduleRoute string) func(http.HandlerFunc, string, ...string) {
	return func(handler http.HandlerFunc, method string, pathVars ...string) {
		path := appendPathVars(fmt.Sprintf("/query/%s/%s", moduleRoute, method), pathVars)
		r.HandleFunc(path, handler).Methods("GET")
	}
}

func appendPathVars(path string, pathVars []string) string {
	for _, v := range pathVars {
		path += fmt.Sprintf("/{%s}", v)
	}
	if len(strings.Fields(path)) > 1 {
		panic(fmt.Errorf("cannot register REST path containing whitespace: %s", path))
	}
	return path
}
