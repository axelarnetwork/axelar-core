package rest

import (
	"bytes"
	"fmt"
	"github.com/axelarnetwork/axelar-core/cmd/abtcd/wallet"
	snapshotRest "github.com/axelarnetwork/axelar-core/x/snapshot/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"io/ioutil"
	"net/http"
)


type RestContext struct {
	URL string
	Codec *codec.Codec
}

func NewRestConext(cdc *codec.Codec, url string) (RestContext) {
	return RestContext {
		Codec: cdc,
		URL: url,
	}
}

func PrepareBaseReq(from string) rest.BaseReq {
	// @TODO Get account information using auth/accounts

	return rest.NewBaseReq(from, "", "axelar", "", "", 3, 1, nil, nil, false)
}

// This is for the account on the val node only
func (rc RestContext) TxSnapshotNow(w wallet.Wallet, from string) error {
	body := snapshotRest.ReqSnapshotNow {
		BaseReq: PrepareBaseReq(from),
	}

	// 1. Build the stdTx using API route
	stdTx, err := rc.RequestTxMessage("tx/snapshot/now", from, body)
	if err != nil {
		return err
	}

	// 2. Sign the tx
	signedStdTx, err := w.SignStdTx(stdTx, false)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", signedStdTx)
	// 3. Broadcast tx
	return nil
}

func (rc RestContext) RequestTxMessage(route string, from string, body interface {}) (auth.StdTx, error){
	json := rc.Codec.MustMarshalJSON(body)
	stdTx := auth.StdTx{}

	resp, err := http.Post(fmt.Sprintf("%s/%s", rc.URL, route), "application/json", bytes.NewBuffer(json))
	if err != nil {
		return stdTx, err
	}
	defer resp.Body.Close()

	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stdTx, err
	}

	err = rc.Codec.UnmarshalJSON(msgBytes, &stdTx)

	return stdTx, err
}