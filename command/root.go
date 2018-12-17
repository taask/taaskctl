package command

import (
	log "github.com/cohix/simplog"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "taaskctl",
		Short: "Taask Core is an open source system for running arbitrary jobs on any infrastructure.",
		Long: `A distributed task execution platform
allowing developers to run intensive and long-running compute tasks
on any infrastructure. Taask is cloud-independent, fully open source,
secure and observable by defult, and runs with zero config in most scenarios`,
		Run: func(cmd *cobra.Command, args []string) {
			log.LogInfo(cmd.Short)
		},
	}
}
