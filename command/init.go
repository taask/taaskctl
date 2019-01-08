package command

import (
	"os"
	"path/filepath"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taask/client-golang"
	"github.com/taask/taask-server/config"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate configuration for taask-server",
		Long: `Init generates the initial configuration needed 
for taask-server to run, and generates the auth configuration that taaskctl needs to connect`,
		Run: func(cmd *cobra.Command, args []string) {
			createConfigDir(taask.DefaultConfigDir())
			createConfigDir(config.DefaultConfigDir())

			adminGroupConfig := taask.GenerateAdminGroup()

			if err := adminGroupConfig.WriteServerConfig(filepath.Join(config.DefaultConfigDir(), "client-auth.yaml")); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteServerConfig"))
				os.Exit(1)
			}

			if err := adminGroupConfig.WriteYAML(filepath.Join(taask.DefaultConfigDir(), "local-auth.yaml")); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteYAML"))
				os.Exit(1)
			}

			log.LogInfo("Taask configured! ✨")
		},
	}
}

func createConfigDir(path string) error {
	return os.MkdirAll(path, 0700)
}
