package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

type ReqSnapshotNow struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

func (r ReqSnapshotNow) GetBaseReq() rest.BaseReq { return r.BaseReq }

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {}
