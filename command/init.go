package command

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
	"github.com/taask/client-golang/config"
	sconfig "github.com/taask/taask-server/config"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate configuration for taask-server",
		Long: `Init generates the initial configuration needed 
for taask-server to run, and generates the auth configuration that taaskctl needs to connect`,
		Run: func(cmd *cobra.Command, args []string) {
			createConfigDir(config.DefaultClientConfigDir())
			createConfigDir(sconfig.DefaultServerConfigDir())
			createConfigDir(DefaultRunnerConfigDir())

			adminGroupConfig := taask.GenerateAdminGroup()

			if err := adminGroupConfig.WriteServerConfig(sconfig.ClientAuthConfigFilename); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteServerConfig"))
				os.Exit(1)
			}

			filename := fmt.Sprintf("%s-auth.yaml", adminGroupConfig.MemberGroup.Name)
			if err := adminGroupConfig.WriteYAML(filepath.Join(config.DefaultClientConfigDir(), filename)); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteYAML"))
				os.Exit(1)
			}

			defaultRunnerGroup := taask.GenerateDefaultRunnerGroup()

			if err := defaultRunnerGroup.WriteServerConfig(sconfig.RunnerAuthConfigFilename); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteServerConfig"))
				os.Exit(1)
			}

			runnerFilename := fmt.Sprintf("%s-auth.yaml", defaultRunnerGroup.MemberGroup.Name)
			if err := defaultRunnerGroup.WriteYAML(filepath.Join(DefaultRunnerConfigDir(), runnerFilename)); err != nil {
				log.LogError(errors.Wrap(err, "failed to WriteYAML"))
				os.Exit(1)
			}

			partnerGroup := taask.GenerateDefaultPartnerGroup()

			if err := partnerGroup.WriteYAML(filepath.Join(sconfig.DefaultServerConfigDir(), sconfig.PartnerAuthConfigFilename)); err != nil {
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

// DefaultRunnerConfigDir returns ~/.taask/runner/config unless XDG_CONFIG_HOME is set
// TODO: find a better place for this
func DefaultRunnerConfigDir() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	root := u.HomeDir
	xdgConfig, useXDG := os.LookupEnv("XDG_CONFIG_HOME")
	if useXDG {
		root = xdgConfig
	}

	return path.Join(root, ".taask/runner/config/")
}
