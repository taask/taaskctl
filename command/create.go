package command

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
	"github.com/taask/taaskctl/readwrite"
)

func createCmd(client *taask.Client) *cobra.Command {
	var watch *bool
	var ugly *bool

	cmd := &cobra.Command{
		Use:   "create [filename | -]",
		Short: "create creates a new task to be run from a file or other input source.",
		Long: `create reads the file indicated by filename (or reads from stdin if - is passed), and creates a task from the input.
The task can be formatted as JSON or YAML.
The task UUID is returned.`,
		Run: func(cmd *cobra.Command, args []string) {
			task, err := readwrite.ReadTaskFile(args[0])
			if err != nil {
				log.LogError(errors.Wrap(err, "failed to ReadTaskFile"))
				os.Exit(1)
			}

			uuid, err := client.SendSpecTask(*task)
			if err != nil {
				log.LogError(errors.Wrap(err, "failed to SendSpecTask"))
				os.Exit(1)
			}

			if *watch {
				result, err := client.StreamTaskResult(uuid)
				if err != nil {
					log.LogError(errors.Wrap(err, "failed to StreamTaskResult"))
					log.LogInfo(fmt.Sprintf("task UUID: %s", uuid))
					os.Exit(1)
				}

				if *ugly {
					fmt.Println(string(result))
					return
				}

				var resultMap map[string]interface{}
				if err := json.Unmarshal(result, &resultMap); err != nil {
					log.LogError(errors.Wrap(err, "failed to Unmarshal result"))
				}

				resultJSON, err := json.MarshalIndent(resultMap, "", "\t")
				if err != nil {
					log.LogError(errors.Wrap(err, "failed to Marshal result"))
					log.LogInfo(fmt.Sprintf("task UUID: %s", uuid))
					os.Exit(1)
				}

				fmt.Println(string(resultJSON))

				return
			}

			fmt.Println(uuid)
		},
	}

	watch = cmd.Flags().Bool("watch", false, "wait for the task result and print it instead of the UUID.")
	ugly = cmd.Flags().Bool("ugly", false, "ugly-print the result JSON. Only applies if combined with --watch.")

	return cmd
}
