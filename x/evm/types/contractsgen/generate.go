package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

//go:generate go run generate.go -template ./contracts.go.tmpl -out ../contracts.go -contracts ../../../../contract-artifacts/contracts

const prefix = "0x"

type contracts struct {
	Token    string
	Burnable string
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

	var buf bytes.Buffer

	if err := t.Execute(&buf, contracts); err != nil {
		log.Fatal(errors.Wrap(err, "failed to apply the contracts template"))
	}

	bz := buf.Bytes()
	bz, err = format.Source(bz)
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not gofmt the output"))
	}

	if err := os.WriteFile(outFP, bz, 0644); err != nil {
		log.Fatal(errors.Wrapf(err, "cannot write to file %s", *out))
	}
}

func parseContracts(contractDir string) contracts {
	var contracts contracts
	contractSetterMapping := map[string]func(string){
		"BurnableMintableCappedERC20": func(bz string) { contracts.Token = strings.TrimPrefix(bz, prefix) },
		"DepositHandler":              func(bz string) { contracts.Burnable = strings.TrimPrefix(bz, prefix) },
	}

	for file, setter := range contractSetterMapping {
		fp, err := filepath.Abs(filepath.Join(contractDir, fmt.Sprintf("%s.sol", file), fmt.Sprintf("%s.json", file)))
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
