package main

import (
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/logging"
)

func main() {
	logger := logging.Logger()

	if err := commands.NewRootCmd().Execute(); err != nil {
		logger.Fatal(err)
	}
}
