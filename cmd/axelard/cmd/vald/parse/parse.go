package parse

import (
	"fmt"
)

// AttributeParser represents a structure to map a single event attribute into an arbitrary data type
type AttributeParser struct {
	Key string
	Map func(string) (interface{}, error)
}

// IdentityMap maps an event attribute into a string without any change
var IdentityMap = func(s string) (interface{}, error) { return s, nil }

// Parse applies all parsers to the given attribute map. Returns map results in the order of the parsers.
func Parse(attributes map[string]string, parsers []*AttributeParser) ([]interface{}, error) {
	results := make([]interface{}, 0, len(parsers))
	for _, parser := range parsers {
		value, ok := attributes[parser.Key]
		if !ok {
			return nil, fmt.Errorf("%s not found", parser.Key)
		}

		result, err := parser.Map(value)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
