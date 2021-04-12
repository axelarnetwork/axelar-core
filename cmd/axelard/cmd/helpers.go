package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/axelarnetwork/axelar-core/utils"
)

func parseThreshold(str string) (utils.Threshold, error) {
	tokens := strings.Split(str, "/")

	if len(tokens) != 2 {
		return utils.Threshold{}, fmt.Errorf("malformed fraction")
	}
	numerator, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return utils.Threshold{}, err
	}
	denominator, err := strconv.ParseInt(tokens[1], 10, 64)
	if err != nil {
		return utils.Threshold{}, err
	}
	return utils.Threshold{Numerator: numerator, Denominator: denominator}, nil
}

func getByteCodes(file string) ([]byte, error) {
	jsonStr, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	table := make(map[string]interface{})
	err = json.Unmarshal(jsonStr, &table)
	if err != nil {
		return nil, err
	}

	str, ok := table["bytecode"].(string)
	if !ok {
		return nil, fmt.Errorf("could not retrieve bytecode from file")
	}

	return hex.DecodeString(str)
}
