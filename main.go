package main

import (
	"os"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/taask/client-golang"
	"github.com/taask/taaskctl/command"
)

func main() {
	client, err := taask.NewClient("localhost", "30688")
	if err != nil {
		log.LogError(errors.Wrap(err, "failed to NewClient"))
		os.Exit(1)
	}

	cmd := command.Build(client)

	if err := cmd.Execute(); err != nil {
		log.LogError(err)
		os.Exit(1)
	}
}
