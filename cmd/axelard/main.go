package main

import (
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd"
)

func main() {
	executor := cmd.NewRootCmd()

	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
