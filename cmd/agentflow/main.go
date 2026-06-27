package main

import (
	"os"

	"github.com/agentflow/agentflow/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
