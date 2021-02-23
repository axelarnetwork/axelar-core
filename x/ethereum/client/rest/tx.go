package rest

import (
	"errors"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"
)

const (
	TxMethodLink               = "link"
	TxMethodVerifyErc20Deploy  = "verify-erc20-deploy"
	TxMethodVerifyErc20Deposit = "verify-erc20-deposit"
	TxMethodSignTx             = "sign-tx"
	TxMethodVerifyTx           = "verify-tx"
	TxMethodSignPending        = "sign-pending"
	TxMethodSignDeployToken    = "sign-deploy-token"

	PathVarChain = "Chain"
	PathVarSymbol = "Symbol"
	PathVarGatewayAddr = "GatewayAddr"
	PathVarTxID = "TxID"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)

	registerTx(GetHandlerLink(cliCtx), TxMethodLink, PathVarChain)
	registerTx(GetHandlerVerifyErc20Deploy(cliCtx), TxMethodVerifyErc20Deploy, PathVarChain)
	registerTx(GetHandlerVerifyErc20Deposit(cliCtx), TxMethodVerifyErc20Deposit, PathVarGatewayAddr, PathVarSymbol)
	registerTx(GetHandlerSignTx(cliCtx), TxMethodSignTx)
	registerTx(GetHandlerVerifyTx(cliCtx), TxMethodVerifyTx)
	registerTx(GetHandlerSignPendingTransfers(cliCtx), TxMethodSignPending)
	registerTx(GetHandlerSignDeployToken(cliCtx), TxMethodSignDeployToken)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryMasterAddress(cliCtx), keeper.QueryMasterAddress)
	registerQuery(QueryCreateDeployTx(cliCtx), keeper.CreateDeployTx)
	registerQuery(QuerySendTx(cliCtx), keeper.SendTx, PathVarTxID)
	registerQuery(QuerySendCommandTx(cliCtx), keeper.SendTx, PathVarGatewayAddr)
}

type ReqLink struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	RecipientAddr string       `json:"recipient" yaml:"recipient"`
	Symbol        string       `json:"symbol" yaml:"symbol"`
	GatewayAddr   string       `json:"gateway_address" yaml:"gateway_address"`
}

type ReqVerifyErc20TokenDeploy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID    string       `json:"tx_id" yaml:"tx_id"`
}

type ReqVerifyErc20Deposit struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID          string       `json:"tx_id" yaml:"tx_id"`
	Amount        string       `json:"amount" yaml:"amount"`
	BurnerAddress string       `json:"burner_address" yaml:"burner_address"`
}

type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxJson  string       `json:"tx_json" yaml:"tx_json"`
}

type ReqSignPendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ReqSignDeployToken struct {
	BaseReq  rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name     string       `json:"name" yaml:"name"`
	Decimals string       `json:"decimals" yaml:"decimals"`
	Capacity string       `json:"capacity" yaml:"capacity"`
}

func GetHandlerLink(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqLink
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.MsgLink{
			Sender:         fromAddr,
			RecipientChain: mux.Vars(r)[PathVarChain],
			RecipientAddr:  req.RecipientAddr,
			Symbol:         req.Symbol,
			GatewayAddr:    req.GatewayAddr,
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerVerifyErc20Deploy(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqVerifyErc20TokenDeploy
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		vars := mux.Vars(r)
		gatewayAddr := common.HexToAddress(vars[PathVarGatewayAddr])
		txID := common.HexToHash(req.TxID)

		msg := types.NewMsgVerifyErc20TokenDeploy(fromAddr, txID, vars[PathVarSymbol], gatewayAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerVerifyErc20Deposit(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqVerifyErc20Deposit
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.NewMsgVerifyErc20Deposit(fromAddr, txID, amount, burnerAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerSignTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		txJson := []byte(req.TxJson)
		var tx *ethTypes.Transaction
		err := cliCtx.Codec.UnmarshalJSON(txJson, &tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSignTx(fromAddr, txJson)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerVerifyTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		json := []byte(req.TxJson)
		err := types.ValidTxJson(json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgVerifyTx(fromAddr, json)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerSignPendingTransfers(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.NewMsgSignPendingTransfers(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerSignDeployToken(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignDeployToken
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		decs, err := strconv.ParseUint(req.Decimals, 10, 8)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse decimals").Error())
		}
		capacity, ok := sdk.NewIntFromString(req.Capacity)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse capacity").Error())
		}

		symbol := mux.Vars(r)[PathVarSymbol]

		msg := types.NewMsgSignDeployToken(fromAddr, req.Name, symbol, uint8(decs), capacity)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
