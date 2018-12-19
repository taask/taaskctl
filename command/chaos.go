package command

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	taask "github.com/taask/client-golang"
)

type addition struct {
	First  int
	Second int
}

type answer struct {
	Answer int
}

func chaosCmd(client *taask.Client) *cobra.Command {
	var numTasks *int

	cmd := &cobra.Command{
		Use:   "chaos",
		Short: "chaos runs load/correctness testing on a Taask installation.",
		Long:  `chaos queues 1000 tasks of Kind io.taask.k8s, waits for them to complete, and prints stats about the run`,
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()

			resultChan := make(chan answer, 2000)

			for i := 0; i < *numTasks; i++ {
				go func(resultChan chan answer) {
					taskBody := addition{
						First:  rand.Intn(50),
						Second: rand.Intn(100),
					}

					taskBodyMap := make(map[string]interface{})
					taskBodyMap["First"] = taskBody.First
					taskBodyMap["Second"] = taskBody.Second

					meta := taask.TaskMeta{
						TimeoutSeconds: 15,
					}

					uuid, err := client.SendTask(taskBodyMap, "io.taask.k8s", meta)
					if err != nil {
						log.LogError(errors.Wrap(err, "failed to SendTask"))
						os.Exit(1)
					}

					resultJSON, err := client.StreamTaskResult(uuid)
					if err != nil {
						log.LogError(errors.Wrap(err, "failed to GetTaskResult"))
						os.Exit(1)
					}

					var taskAnswer answer
					if err := json.Unmarshal(resultJSON, &taskAnswer); err != nil {
						log.LogError(errors.Wrap(err, "failed to Unmarshal"))
					}

					resultChan <- taskAnswer
				}(resultChan)
			}

			completed := 0
			log.LogInfo("waiting for answers")

			for {
				answer := <-resultChan
				log.LogInfo(fmt.Sprintf("task answer: %d", answer.Answer))

				completed++

				log.LogInfo(fmt.Sprintf("%d/%d completed", completed, *numTasks))

				if completed == *numTasks {
					break
				}
			}

			duration := time.Since(start)
			log.LogInfo(fmt.Sprintf("took %s", duration.String()))
		},
	}

	numTasks = cmd.Flags().Int("count", 1000, "the number of tasks to execute")

	return cmd
}
