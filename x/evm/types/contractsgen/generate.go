package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

//go:generate go run generate.go -template ./contracts.go.tmpl -out ../contracts.go -contracts ../../../../contract-artifacts/gateway

type contracts struct {
	SinglesigGateway string
	MultisigGateway  string
	Token            string
	Burnable         string
}

func main() {
	contractDir := flag.String("contracts", "./", "contracts directory")
	out := flag.String("out", "./contracts.go", "output location")
	templateFP := flag.String("template", "./contracts.go.tmpl", "template location")

	flag.Parse()

	contracts := parseContracts(*contractDir)

	t := template.Must(template.ParseFiles(*templateFP))

	outFP, err := filepath.Abs(*out)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "cannot build filepath for %s", *out))
	}

	f, err := os.OpenFile(outFP, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "cannot open or create file %s", *out))
	}

	if err := t.Execute(f, contracts); err != nil {
		log.Fatal(errors.Wrap(err, "failed to apply the contracts template"))
	}

	if err := f.Close(); err != nil {
		log.Fatal(errors.Wrapf(err, "cannot close file %s", *out))
	}
}

func parseContracts(contractDir string) contracts {
	var contracts contracts
	contractSetterMapping := map[string]func(string){
		"AxelarGatewayProxySinglesig": func(bz string) { contracts.SinglesigGateway = bz },
		"AxelarGatewayProxyMultisig":  func(bz string) { contracts.MultisigGateway = bz },
		"BurnableMintableCappedERC20": func(bz string) { contracts.Token = bz },
		"DepositHandler":              func(bz string) { contracts.Burnable = bz },
	}

	for file, setter := range contractSetterMapping {
		fp, err := filepath.Abs(filepath.Join(contractDir, file+".json"))
		if err != nil {
			log.Fatal(errors.Wrapf(err, "cannot build filepath for %s", file))
		}

		content, err := os.ReadFile(fp)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "failed to read contract %s", file))
		}

		jsonMap := make(map[string]interface{})
		if err := json.Unmarshal(content, &jsonMap); err != nil {
			log.Fatal(errors.Wrapf(err, "failed to json parse contract %s", file))
		}
		setter(jsonMap["bytecode"].(string))
	}
	return contracts
}
