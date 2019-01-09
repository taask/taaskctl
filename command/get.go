package command

import (
	"fmt"
	"os"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
	"github.com/taask/taask-server/model"
)

func getCmd(client *taask.Client) *cobra.Command {
	var watch *bool
	var ugly *bool

	cmd := &cobra.Command{
		Use:   "get [uuid]",
		Short: "gets the results of a task.",
		Long: `get fetches the status of task [uuid].
If the task is not complete, the task's status is printed.
If the task is complete, the result JSON is printed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if client == nil {
				log.LogError(errors.New("unable to connect"))
				return
			}

			uuid := args[0]

			if *watch {
				watchResult(client, uuid, *ugly)
				return
			}

			status, err := client.GetTaskStatus(uuid)
			if err != nil {
				log.LogError(errors.Wrap(err, "failed to GetTaskStatus"))
				os.Exit(1)
			}

			if status == model.TaskStatusCompleted {
				watchResult(client, uuid, *ugly) // this will just print the result
				return
			}

			log.LogInfo(fmt.Sprintf("task %s status %s", uuid, status))
		},
	}

	watch = cmd.Flags().Bool("watch", false, "wait for the task result and print it instead of the UUID.")
	ugly = cmd.Flags().Bool("ugly", false, "ugly-print the result JSON. Only applies if combined with --watch.")

	return cmd
}
