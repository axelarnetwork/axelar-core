package rest

import (
	"fmt"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesRest "github.com/cosmos/cosmos-sdk/types/rest"
	typesAuth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"io/ioutil"
	"net/http"
)

type  QueryRespAccount struct {
	Height uint64 `json:"height" yaml:"height"`
	Account typesAuth.BaseAccount `json:"result" yaml:"result"`
}


// @TODO abstract
func (rc RestContext) RequestQuery(restRoute string) (queryResp typesRest.ResponseWithHeight, err error){

	uri := fmt.Sprintf("%s/%s", rc.URL, restRoute)
	resp, err := http.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msgBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("GET %s: Error%+v\n", uri, string(msgBytes))
		err = errors.New(fmt.Sprintf("Query to %s resulted in status %s", uri, resp.Status))
		return
	}

	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = rc.Codec.UnmarshalJSON(msgBytes, &queryResp)

	return
}

func (rc RestContext) QueryAccount(address sdk.AccAddress) (account typesAuth.BaseAccount, err error){

	uri := fmt.Sprintf("%s/auth/accounts/%s", rc.URL, address.String())
	resp, err := http.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msgBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("GET %s: Error%+v\n", uri, string(msgBytes))
		err = errors.New(fmt.Sprintf("Query to %s resulted in status %s", uri, resp.Status))
		return
	}

	msgBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var queryResp typesRest.ResponseWithHeight
	err = rc.Codec.UnmarshalJSON(msgBytes, &queryResp)
	if err != nil {
		return
	}

	err = rc.Codec.UnmarshalJSON(queryResp.Result, &account)
	if err != nil {
		return
	}

	return
}
