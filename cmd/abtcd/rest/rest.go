package rest

import (
	"bytes"
	"fmt"
	"errors"
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

func prepareBaseReq(w wallet.Wallet) rest.BaseReq {
	// @TODO Get account information using auth/accounts

	//return rest.NewBaseReq(string(w.FromAddr), "", w.Config.ChainID, w.Config.Gas, w.Config.GasAdjustment., 3, 1, w.Config.GasFees, w.Config.GasPrices, false)
	return rest.NewBaseReq(string(w.FromAddr), "", w.Config.ChainID, w.Config.Gas, "", 3, 1, w.Config.GasFees, w.Config.GasPrices, false)
}

// This is for the account on the val node only
// @TODO abstract into func SubmitTx(req, route string)
func (rc RestContext) TxSnapshotNow(w wallet.Wallet) error {
	body := snapshotRest.ReqSnapshotNow {
		BaseReq: prepareBaseReq(w),
	}

	// 0. Update account nonce

	// 1. Build the stdTx using API route
	stdTx, err := rc.RequestTxMessage("tx/snapshot/now", body)
	if err != nil {
		return err
	}

	// 2. Sign the tx
	signedStdTx, err := w.SignStdTx(stdTx, false)
	if err != nil {
		return err
	}

	fmt.Printf("signedTx pub keys: %+v\n", signedStdTx.GetPubKeys())
	// 3. Broadcast tx
	if err := rc.BroadcastSignedTx(signedStdTx); err != nil {
		return err
	}

	return nil
}

func (rc RestContext) RequestTxMessage(route string, body interface {}) (auth.StdTx, error){
	json := rc.Codec.MustMarshalJSON(body)
	stdTx := auth.StdTx{}

	uri := fmt.Sprintf("%s/%s", rc.URL, route)
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(json))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return stdTx, err
	}

	if resp.StatusCode != 200 {
		return stdTx, errors.New(fmt.Sprintf("Post to %s resulted in status %s", uri, resp.Status))
	}


	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stdTx, err
	}

	err = rc.Codec.UnmarshalJSON(msgBytes, &stdTx)

	fmt.Printf("Response from %s:\n%+v\n", uri, string(msgBytes))
	fmt.Printf("Unmarshalled response from %s:\n%+v\n", uri, stdTx)
	return stdTx, err
}

func (rc RestContext) BroadcastSignedTx(stdTx auth.StdTx) error {
	json := rc.Codec.MustMarshalJSON(stdTx)
	uri := fmt.Sprintf("%s/%s", rc.URL, "txs")

	fmt.Printf("Signed msg json %s:\n%+v\n", uri, string(json))
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(json))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Post to %s resulted in status %s", uri, resp.Status))
	}

	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//err = rc.Codec.UnmarshalJSON(msgBytes, &stdTx)
	fmt.Printf("Unmarshalled response from %s:\n%+v\n", uri, string(msgBytes))
	return nil
}