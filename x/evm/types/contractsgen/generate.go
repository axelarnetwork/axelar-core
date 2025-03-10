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
		log.Fatal(fmt.Errorf("cannot build filepath for %s: %w", *out, err))
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, contracts); err != nil {
		log.Fatal(fmt.Errorf("failed to apply the contracts template: %w", err))
	}

	bz := buf.Bytes()
	bz, err = format.Source(bz)
	if err != nil {
		log.Fatal(fmt.Errorf("could not gofmt the output: %w", err))
	}

	if err := os.WriteFile(outFP, bz, 0644); err != nil {
		log.Fatal(fmt.Errorf("cannot write to file %s: %w", *out, err))
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
			log.Fatal(fmt.Errorf("cannot build filepath for %s: %w", file, err))
		}

		content, err := os.ReadFile(fp)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to read contract %s: %w", file, err))
		}

		jsonMap := make(map[string]interface{})
		if err := json.Unmarshal(content, &jsonMap); err != nil {
			log.Fatal(fmt.Errorf("failed to json parse contract %s: %w", file, err))
		}
		setter(jsonMap["bytecode"].(string))
	}
	return contracts
}
