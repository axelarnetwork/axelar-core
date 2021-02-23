package rest

import (
	"github.com/axelarnetwork/axelar-core/utils/denom"
	"github.com/btcsuite/btcutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	TxMethodLink                   = "link"
	TxMethodVerifyTx               = "verify"
	TxMethodSignPendingTransfersTx = "sign"

	QMethodDepositAddress = keeper.QueryDepositAddress
	QMethodTxInfo         = keeper.QueryOutInfo
	QMethodSendTransfers  = keeper.SendTx

	PathVarChain           = "Chain"
	PathVarEthereumAddress = "EthereumAddress"
	PathVarTxID            = "TxID"
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerLink(cliCtx), TxMethodLink, PathVarChain)
	registerTx(GetHandlerVerifyTx(cliCtx), TxMethodVerifyTx)
	registerTx(GetHandlerSignPendingTransfersTx(cliCtx), TxMethodSignPendingTransfersTx)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryDepositAddress(cliCtx), QMethodDepositAddress, PathVarChain, PathVarEthereumAddress)
	registerQuery(QueryTxInfo(cliCtx), QMethodTxInfo, PathVarTxID)
	registerQuery(QuerySendTransfers(cliCtx), QMethodSendTransfers)
}

// ReqLink represents a request to link a cross-chain address to a Bitcoin address
type ReqLink struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Address string       `json:"address" yaml:"address"`
}

// ReqVerifyTx represents a request to verify a Bitcoin transaction
type ReqVerifyTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxInfo  string       `json:"tx_info" yaml:"tx_info"`
}

// ReqSignPendingTransfersTx represents a request to sign pending token transfers from Ethereum
type ReqSignPendingTransfersTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Fee     string       `json:"fee" yaml:"fee"`
}

func GetHandlerSignPendingTransfersTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignPendingTransfersTx
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

		satoshi, err := denom.ParseSatoshi(req.Fee)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSign(fromAddr, btcutil.Amount(satoshi.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
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

		msg := types.MsgLink{Sender: fromAddr, RecipientChain: mux.Vars(r)[PathVarChain], RecipientAddr: req.Address}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerVerifyTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqVerifyTx
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

		var out types.OutPointInfo
		if err := cliCtx.Codec.UnmarshalJSON([]byte(req.TxInfo), &out); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.MsgVerifyTx{Sender: fromAddr, OutPointInfo: out}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
