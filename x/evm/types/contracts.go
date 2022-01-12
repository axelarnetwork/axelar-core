package types

import (
	"io/ioutil"
	"encoding/json"
)

func getBytecodeFromArtifact(filename string) string {
    var compiled map[string]interface{}
    byteValue, err := ioutil.ReadFile("../../../artifacts/" + filename)

    if err != nil {
        panic(err)
    }

    json.Unmarshal(byteValue, &compiled)

    return compiled["bytecode"].(string)
}

// TODO: use templating for this file instead of dynamic import
// so imported bytecode will be a part of go binary
var (
	singlesigGateway = getBytecodeFromArtifact("AxelarGatewayProxySinglesig.json")
	multisigGateway  = getBytecodeFromArtifact("AxelarGatewayProxyMultisig.json")
	token            = getBytecodeFromArtifact("BurnableMintableCappedERC20.json")
	burnable         = getBytecodeFromArtifact("Burner.json")
)
