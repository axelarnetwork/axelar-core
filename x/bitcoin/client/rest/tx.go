package rest

import (
	"fmt"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router ) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/link/{chain}", types.ModuleName), linkHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/track/pubkey", types.ModuleName), trackPubKeyHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/track/address/{address}", types.ModuleName), trackAddressHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/verify", types.ModuleName), verifyTxHandlerFn(cliCtx)).Methods("POST")
	//r.HandleFunc(fmt.Sprintf("/tx/%s/sign/{txId}", types.ModuleName), signTxHandlerFn(cliCtx)).Methods("POST")

	//r.HandleFunc(fmt.Sprintf("/query/%s/%s", types.ModuleName, keeper.QueryMasterAddress), QueryMasterAddressHandlerFn(cliCtx, keeper.QueryMasterAddress)).Methods("GET")
}

type ReqTrackPubKey struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	UseMasterKey bool `json:"use_master_key" yaml:"master_key"`
	KeyId string `json:"key_id" yaml:"key_id"`
	Rescan bool `json:"rescan" yaml:"rescan"`
}

type ReqTrackAddress struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Rescan bool `json:"rescan" yaml:"rescan"`
}

type ReqLink struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Address string `json:"key_id" yaml:"key_id"`
}

type ReqVerifyTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxInfo string `json:"tx_info" yaml:"rescan"`
}

type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxJson string `json:"tx_json" yaml:"rescan"`
}

func linkHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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

		vars := mux.Vars(r)
		chain := balance.ChainFromString(vars["chain"])
		address := balance.CrossChainAddress{Chain: chain, Address: req.Address}
		if err := address.Validate(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		msg := types.MsgLink{Sender: fromAddr, Recipient: address}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func trackAddressHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var req ReqTrackAddress
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
		addr, err := btcutil.DecodeAddress(vars["address"], &chaincfg.MainNetParams)
		if err != nil {
			return
		}

		msg := types.NewMsgTrackAddress(fromAddr, addr.String(), req.Rescan)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func trackPubKeyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var req ReqTrackPubKey
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

		var msg sdk.Msg
		if (req.UseMasterKey && req.KeyId != "") || (!req.UseMasterKey && req.KeyId == "") {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "either set the flag to use a key ID or to use the master key, not both")
		}

		msg = types.NewMsgTrackAddress(fromAddr, req.KeyId, req.Rescan)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func verifyTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

//func signTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		var req ReqSignTx
//		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
//			return
//		}
//		req.BaseReq = req.BaseReq.Sanitize()
//		if !req.BaseReq.ValidateBasic(w) {
//			return
//		}
//		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
//		if !ok {
//			return
//		}
//
//		vars := mux.Vars(r)
//		var tx *wire.MsgTx
//		if err := types.ModuleCdc.UnmarshalJSON([]byte(req.TxJson), &tx); err != nil {
//			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
//			return
//		}
//
//		msg := types.NewMsgSignTx(fromAddr, vars["txId"], tx)
//
//		if err := msg.ValidateBasic(); err != nil {
//			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
//			return
//		}
//		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
//	}
//}