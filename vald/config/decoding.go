package config

import (
	"reflect"

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

// AddDecodeHooks adds decode hooks to the given config to correctly translate string into FinalityOverride
func AddDecodeHooks(cfg *mapstructure.DecoderConfig) {
	hooks := []mapstructure.DecodeHookFunc{
		stringToFinalityOverride,
	}
	if cfg.DecodeHook != nil {
		hooks = append(hooks, cfg.DecodeHook)
	}

	cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(hooks...)
}
