package command

import (
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
)

// Build builds the command tree
func Build(client *taask.Client) *cobra.Command {
	root := rootCmd()

	// Generate auth for deploying taask
	root.AddCommand(initCmd())

	// Task commands
	root.AddCommand(createCmd(client))
	root.AddCommand(getCmd(client))

	// Load testing
	root.AddCommand(chaosCmd(client))

	return root
}
