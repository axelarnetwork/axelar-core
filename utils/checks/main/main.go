package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils/checks"
)

func main() {
	cmd := &cobra.Command{
		Use: "check",
	}
	cmd.AddCommand(checks.FieldDeclarations())

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
