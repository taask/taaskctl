package main

import (
	"os"
	"path/filepath"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/taask/client-golang"
	"github.com/taask/taaskctl/command"
)

func main() {
	client, err := createClient()
	if err != nil {
		// TODO: figure out a better place to create the client
		// log.LogWarn(errors.Wrap(err, "failed to createClient").Error())
	}

	cmd := command.Build(client)

	if err := cmd.Execute(); err != nil {
		log.LogError(err)
		os.Exit(1)
	}
}

func createClient() (*taask.Client, error) {
	authPath := filepath.Join(taask.DefaultConfigDir(), "local-auth.yaml")

	localAuthConfig, err := taask.LocalAuthConfigFromFile(authPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to LocalAuthConfigFromFile")
	}

	client, err := taask.NewClient("localhost", "30688", localAuthConfig)
	if err != nil {
		log.LogWarn(errors.Wrap(err, "failed to NewClient").Error())
	}

	return client, nil
}
