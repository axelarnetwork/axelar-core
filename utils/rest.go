package utils

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

const (
	PathVarChain           = "Chain"
	PathVarKeyRole         = "KeyRole"
	PathVarEthereumAddress = "EthereumAddress"
	PathVarSymbol          = "Symbol"
	PathVarTxID            = "TxID"
	PathVarCommandID       = "CommandID"
)

// Extract the sender address from an SDK base request
func ExtractReqSender(w http.ResponseWriter, req rest.BaseReq) (sdk.AccAddress, bool) {
	sender, err := sdk.AccAddressFromBech32(req.From)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return sender, true
}

func RegisterTxHandlerFn(r *mux.Router, moduleRoute string) func(http.HandlerFunc, string, ...string) {
	return func(handler http.HandlerFunc, method string, pathVars ...string) {
		path := appendPathVars(fmt.Sprintf("/tx/%s/%s", moduleRoute, method), pathVars)
		r.HandleFunc(path, handler).Methods("POST")
	}
}

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
