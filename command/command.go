package command

import (
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
)

// Build builds the command tree
func Build(client *taask.Client) *cobra.Command {
	root := rootCmd()

	root.AddCommand(chaosCmd(client))
	root.AddCommand(createCmd(client))

	return root
}
