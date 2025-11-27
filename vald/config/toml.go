package config

import (
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mitchellh/mapstructure"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
)

// WriteTOML encodes the given config into TOML and writes it to the given io.Writer.
// It uses mapstructure tags to determine the TOML keys.
func WriteTOML(w io.Writer, cfg interface{}) error {
	mapped, err := structToMap(cfg)
	if err != nil {
		return err
	}

	return toml.NewEncoder(w).Encode(mapped)
}

// structToMap converts a struct to a map using mapstructure tags for keys.
// Unlike mapstructure.Decode, this recursively converts nested structs and slices.
func structToMap(cfg interface{}) (map[string]interface{}, error) {
	mapped := make(map[string]interface{})

	decoderConfig := &mapstructure.DecoderConfig{
		Result: &mapped,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	// Recursively convert any nested structs/slices that mapstructure didn't convert
	for k, v := range mapped {
		converted, err := convertValue(v)
		if err != nil {
			return nil, err
		}
		mapped[k] = converted
	}

	return mapped, nil
}

// convertValue recursively converts structs and slices to maps/slices of maps.
// Special types like time.Duration and sdk.AccAddress are converted to strings.
func convertValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	// Handle special types before they lose type information
	switch d := v.(type) {
	case time.Duration:
		return d.String(), nil
	case sdk.AccAddress:
		if d == nil {
			return "", nil
		}
		return d.String(), nil

	case rpc.FinalityOverride:
		switch d {
		case rpc.NoOverride:
			return nil, nil
		case rpc.Confirmation:
			return strings.ToLower(d.String()), nil
		}
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Struct:
		// Convert struct to map using mapstructure
		return structToMap(v)
	case reflect.Slice:
		// Convert each element
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			converted, err := convertValue(elem)
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}
		return result, nil
	case reflect.Map:
		// Recursively convert map values
		result := make(map[string]interface{})
		for iter := val.MapRange(); iter.Next(); {
			k := iter.Key().String()
			converted, err := convertValue(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			result[k] = converted
		}
		return result, nil
	default:
		return v, nil
	}
}
