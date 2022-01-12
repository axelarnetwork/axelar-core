package types

import (
	"io/ioutil"
	"encoding/json"
)

func getBytecodeFromArtifact(filename string) string {
    var compiled map[string]interface{}
    byteValue, _ := ioutil.ReadFile("artifacts/" + filename)

    json.Unmarshal(byteValue, &compiled)

    return compiled["bytecode"].(string)
}

var (
	singlesigGateway = getBytecodeFromArtifact("AxelarGatewaySinglesig.json")
	multisigGateway  = getBytecodeFromArtifact("AxelarGatewaySinglesig.json")
	token            = getBytecodeFromArtifact("ERC20.json")
	burnable         = getBytecodeFromArtifact("Burner.json")
)
