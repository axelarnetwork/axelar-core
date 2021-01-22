package rest

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

type RestContext struct {
	Codec *codec.Codec
	URL   string
}

func NewRestConext(cdc *codec.Codec, url string) RestContext {
	return RestContext{
		Codec: cdc,
		URL:   url,
	}
}
