package config

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mitchellh/mapstructure"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
)

func stringToFinalityOverride(
	f reflect.Type,
	t reflect.Type,
	data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf(rpc.FinalityOverride(0)) {
		return data, nil
	}

	return rpc.ParseFinalityOverride(data.(string))
}

func stringToAccAddress(
	f reflect.Type,
	t reflect.Type,
	data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(sdk.AccAddress{}) {
		return data, nil
	}

	s := data.(string)
	if s == "" {
		return sdk.AccAddress(nil), nil
	}
	return sdk.AccAddressFromBech32(s)
}

// AddDecodeHooks adds decode hooks to correctly translate string types
func AddDecodeHooks(cfg *mapstructure.DecoderConfig) {
	hooks := []mapstructure.DecodeHookFunc{
		stringToFinalityOverride,
		stringToAccAddress,
	}
	if cfg.DecodeHook != nil {
		hooks = append(hooks, cfg.DecodeHook)
	}

	cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(hooks...)
}
