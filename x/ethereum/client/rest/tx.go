package rest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

// rest routes
const (
	TxMethodLink               = "link"
	TxMethodConfirmTokenDeploy = "confirm-erc20-deploy"
	TxMethodConfirmDeposit     = "confirm-erc20-deposit"
	TxMethodSignTx             = "sign-tx"
	TxMethodSignPending        = "sign-pending"
	TxMethodSignDeployToken    = "sign-deploy-token"
	TxMethodSignBurnTokens     = "sign-burn"
	TxAddChain                 = "add-chain"

	QueryMethodMasterAddress        = keeper.QueryMasterAddress
	QueryMethodAxelarGatewayAddress = keeper.QueryAxelarGatewayAddress
	QueryMethodCommandData          = keeper.QueryCommandData
	QueryMethodCreateDeployTx       = keeper.CreateDeployTx
	QueryMethodSendTx               = keeper.SendTx
	QueryMethodSendCommand          = keeper.SendCommand
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerLink(cliCtx), TxMethodLink, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTokenDeploy(cliCtx), TxMethodConfirmTokenDeploy, clientUtils.PathVarSymbol)
	registerTx(GetHandlerConfirmDeposit(cliCtx), TxMethodConfirmDeposit)
	registerTx(GetHandlerSignTx(cliCtx), TxMethodSignTx)
	registerTx(GetHandlerSignPendingTransfers(cliCtx), TxMethodSignPending)
	registerTx(GetHandlerSignDeployToken(cliCtx), TxMethodSignDeployToken, clientUtils.PathVarSymbol)
	registerTx(GetHandlerSignBurnTokens(cliCtx), TxMethodSignBurnTokens)
	registerTx(GetHandlerAddChain(cliCtx), TxAddChain)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(GetHandlerQueryMasterAddress(cliCtx), QueryMethodMasterAddress)
	registerQuery(GetHandlerQueryAxelarGatewayAddress(cliCtx), QueryMethodAxelarGatewayAddress)
	registerQuery(GetHandlerQueryCommandData(cliCtx), QueryMethodCommandData, clientUtils.PathVarCommandID)
	registerQuery(GetHandlerQueryCreateDeployTx(cliCtx), QueryMethodCreateDeployTx)
	registerQuery(GetHandlerQuerySendTx(cliCtx), QueryMethodSendTx, clientUtils.PathVarTxID)
	registerQuery(GetHandlerQuerySendCommandTx(cliCtx), QueryMethodSendCommand)
}

// ReqLink represents a request to link a cross-chain address to an Ethereum address
type ReqLink struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	RecipientAddr string       `json:"recipient" yaml:"recipient"`
	Symbol        string       `json:"symbol" yaml:"symbol"`
}

// ReqConfirmTokenDeploy represents a request to confirm a token deployment
type ReqConfirmTokenDeploy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID    string       `json:"tx_id" yaml:"tx_id"`
}

// ReqConfirmDeposit represents a request to confirm a deposit
type ReqConfirmDeposit struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID          string       `json:"tx_id" yaml:"tx_id"`
	Amount        string       `json:"amount" yaml:"amount"`
	BurnerAddress string       `json:"burner_address" yaml:"burner_address"`
}

// ReqSignTx represents a request to sign a transaction
type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxJSON  string       `json:"tx_json" yaml:"tx_json"`
}

// ReqSignPendingTransfers represents a request to sign all pending transfers
type ReqSignPendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqSignDeployToken represents a request to sign a deploy token command
type ReqSignDeployToken struct {
	BaseReq  rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name     string       `json:"name" yaml:"name"`
	Decimals string       `json:"decimals" yaml:"decimals"`
	Capacity string       `json:"capacity" yaml:"capacity"`
}

// ReqSignBurnTokens represents a request to sign all outstanding burn commands
type ReqSignBurnTokens struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqAddchain represents a request to add a new evm chain command
type ReqAddchain struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name            string       `json:"name" yaml:"name"`
	NativeAsset     string       `json:"native_asset" yaml:"native_asset"`
	SupportsForeign bool         `json:"supports_foreign" yaml:"supports_foreign"`
}

// GetHandlerLink returns the handler to link addresses
func GetHandlerLink(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqLink
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := &types.LinkRequest{
			Sender:         fromAddr,
			RecipientChain: mux.Vars(r)[clientUtils.PathVarChain],
			RecipientAddr:  req.RecipientAddr,
			Symbol:         req.Symbol,
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmTokenDeploy returns a handler to confirm a token deployment
func GetHandlerConfirmTokenDeploy(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmTokenDeploy
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		txID := common.HexToHash(req.TxID)
		msg := types.NewConfirmTokenRequest(fromAddr, txID, mux.Vars(r)[clientUtils.PathVarSymbol])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmDeposit returns a handler to confirm a deposit
func GetHandlerConfirmDeposit(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmDeposit
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		txID := common.HexToHash(req.TxID)
		amount := sdk.NewUintFromString(req.Amount)
		burnerAddr := common.HexToAddress(req.BurnerAddress)

		msg := types.NewConfirmDepositRequest(fromAddr, txID, amount, burnerAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignTx returns a handler to sign a transaction
func GetHandlerSignTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		txJSON := []byte(req.TxJSON)
		var ethtx *ethTypes.Transaction
		err := cliCtx.LegacyAmino.UnmarshalJSON(txJSON, &ethtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewSignTxRequest(fromAddr, txJSON)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignPendingTransfers returns a handler to sign all pending transfers
func GetHandlerSignPendingTransfers(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewSignPendingTransfersRequest(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignDeployToken returns a handler to sign a deploy token command
func GetHandlerSignDeployToken(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignDeployToken
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		symbol := mux.Vars(r)[clientUtils.PathVarSymbol]
		decs, err := strconv.ParseUint(req.Decimals, 10, 8)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse decimals").Error())
		}
		capacity, ok := sdk.NewIntFromString(req.Capacity)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse capacity").Error())
		}

		msg := types.NewSignDeployTokenRequest(fromAddr, req.Name, symbol, uint8(decs), capacity)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignBurnTokens returns a handler to sign all outstanding burn commands
func GetHandlerSignBurnTokens(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignBurnTokens
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewSignBurnTokensRequest(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerAddChain returns a handler to add a new evm chain command
func GetHandlerAddChain(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqAddchain
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewAddChainRequest(fromAddr, req.Name, req.NativeAsset, req.SupportsForeign)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
