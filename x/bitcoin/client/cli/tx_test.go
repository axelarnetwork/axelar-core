package cli

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
)

func TestUnmarshalGetTxOut(t *testing.T) {
	j, err := ioutil.ReadFile("./testdata/txout.json")
	if err != nil {
		panic(err)
	}

	var result btcjson.GetTxOutResult
	err = json.Unmarshal(j, &result)
	if err != nil {
		panic(err)
	}
}
