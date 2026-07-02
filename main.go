package main

import (
	"os"

	"autodeploy/internal/modules/deploy"
)

func main() {
	cmd := deploy.NewDeployCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
