package rest

import (
	"io/ioutil"
	"net/http"
	"bytes"
	"fmt"
	"errors"

	"github.com/axelarnetwork/axelar-core/cmd/abtcd/wallet"
	snapshotRest "github.com/axelarnetwork/axelar-core/x/snapshot/client/rest"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	typesRest "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
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

func prepareBaseReq(w wallet.Wallet) typesRest.BaseReq {
	// @TODO Get account information using auth/accounts

	// @NB "signer index" in broadcast handler refers to signer pub keys slice on signed stdTx
	return typesRest.NewBaseReq(w.FromAddr.String(), "", w.Config.ChainID, w.Config.Gas, "", w.AccountNumber, w.SequenceNumber, w.Config.GasFees, w.Config.GasPrices, false)
}

// This is for the account on the val node only
// @TODO abstract into func SubmitTx(req, route string)
func (rc RestContext) TxSnapshotNow(w wallet.Wallet) error {
	// 0. Update account nonce

	// 1. Build the stdTx using API route
	body := snapshotRest.ReqSnapshotNow {
		BaseReq: prepareBaseReq(w),
	}

	txRoute := "tx/snapshot/now"
	stdTx, err := rc.RequestTxMessage(txRoute, body)
	if err != nil {
		return err
	}

	// 2. Sign the tx
	signedStdTx, err := w.SignStdTx(stdTx, false)
	if err != nil {
		return err
	}

	// 3. Broadcast tx
	txResp, err := rc.BroadcastSignedTx(signedStdTx, "block")
	if err != nil {
		return err
	}

	if !TxResponseSuccess(txResp) {
		reason := txResp.RawLog
		if len(reason) == 0 { reason = fmt.Sprintf("Code %d", txResp.Code) }
		return errors.New(fmt.Sprintf("TxSubmission to %s failed. Reason: %s", txRoute, reason))
	}
	fmt.Printf("Tx SUCCESS: %s at height %d with txHash: %s\n\n", txRoute, txResp.Height, txResp.TxHash)

	return nil
}


func (rc RestContext) RequestTxMessage(route string, body interface {}) (auth.StdTx, error){
	json := rc.Codec.MustMarshalJSON(body)
	stdTx := auth.StdTx{}

	uri := fmt.Sprintf("%s/%s", rc.URL, route)
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(json))
	if err != nil {

		return stdTx, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msgBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("POST %s: Error%+v\n", uri, string(msgBytes))
		return stdTx, errors.New(fmt.Sprintf("Post to %s resulted in status %s", uri, resp.Status))
	}


	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stdTx, err
	}

	err = rc.Codec.UnmarshalJSON(msgBytes, &stdTx)

	fmt.Printf("From %s:\n%+v\n\n", uri, string(msgBytes))
	//fmt.Printf("From %s: Unmarshalled stdTx\n%+v\n", uri, stdTx)
	return stdTx, err
}

func (rc RestContext) BroadcastSignedTx(stdTx auth.StdTx, mode string) (txResp sdk.TxResponse, err error) {
	broadcastReq := authRest.BroadcastReq{
		Tx:   stdTx,
		Mode: mode,
	}
	json := rc.Codec.MustMarshalJSON(broadcastReq)
	uri := fmt.Sprintf("%s/%s", rc.URL, "txs")

	fmt.Printf("From %s: Signed msg json\n%+v\n\n", uri, string(json))
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(json))
	if resp != nil {
		defer func() {
			if rerr := resp.Body.Close(); rerr != nil && err == nil {
				err = rerr
			}
		}()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		msgBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("POST %s: Error%+v\n", uri, string(msgBytes))
		err = errors.New(fmt.Sprintf("Post to %s resulted in status %s", uri, resp.Status))
		return
	}

	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = rc.Codec.UnmarshalJSON(msgBytes, &txResp); err != nil {
		return
	}

	success := TxResponseSuccess(txResp)
	if !success {
		fmt.Printf("Tx FAILURE: broadcast failed with code %s#%d:\n%+v\n", txResp.Codespace, txResp.Code, string(msgBytes))
	}
	return
}

func TxResponseSuccess(resp sdk.TxResponse) bool {
	if resp.Height == 0 || resp.Empty() {
		return false
	}

	return true
}