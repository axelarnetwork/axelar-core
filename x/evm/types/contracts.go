package types

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

func getBytecodeFromArtifact(filename string) string {
    var compiled map[string]interface{}
    byteValue, err := ioutil.ReadFile("../../../artifacts/" + filename)

    if err != nil {
        path, _ := os.Getwd()
        fmt.Println(path)
        panic(err)
    }

    json.Unmarshal(byteValue, &compiled)

    return compiled["bytecode"].(string)
}

var (
	singlesigGateway = getBytecodeFromArtifact("AxelarGatewaySinglesig.json")
	multisigGateway  = getBytecodeFromArtifact("AxelarGatewaySinglesig.json")
	token            = getBytecodeFromArtifact("ERC20.json")
	burnable         = getBytecodeFromArtifact("Burner.json")
)
