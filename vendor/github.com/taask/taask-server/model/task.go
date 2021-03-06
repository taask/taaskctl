// this file contains extensions to the Task struct generated by protoc, do not edit the associated .pb.go file.

package model

import (
	"fmt"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
)

// TaskStatusWaiting and others represent the status of a task
const (
	// the task has been recieved, but the client does not want it to be scheduled yet.
	TaskStatusPending = "pending"

	// task is waiting when it has been received but not yet scheduled to a runner
	// waiting tasks will be queued at the first available opportunity
	TaskStatusWaiting = "waiting"

	// task goes waiting -> queued when it has been scheduled to a runner
	TaskStatusQueued = "queued"

	// task goes queued -> running when the runner has begun executing it
	TaskStatusRunning = "running"

	// task goes running -> complete if the runner completes the task with no critical errors.
	// complete does not indicate wether the result of the job contained an error or not, but rather that the runner was able to finish running the task
	TaskStatusCompleted = "complete"

	// task goes running -> failed if the runner cannot complete the job for whatever reason (crash, etc.)
	TaskStatusFailed = "failed"

	// task goes queued -> retrying if the runner a task was scheduled to fails to run the task
	// task goes running -> retrying if the task (runs past the deadline OR runner dies) AND is marked as retryable
	// retrying is similar to waiting, except there will be a backoff before it is re-queued
	TaskStatusRetrying = "retrying"
)

// Update applies an update to a task object and returns the update with the updated version number
func (t *Task) Update(u TaskUpdate) (TaskUpdate, error) {
	u.UUID = t.UUID
	u.Version = t.Meta.Version + 1

	if err := t.ApplyUpdate(u, false); err != nil {
		return TaskUpdate{}, errors.Wrap(err, "failed to ApplyUpdate")
	}

	return u, nil
}

// ApplyUpdate applies an update to a task
func (t *Task) ApplyUpdate(update TaskUpdate, logIt bool) error {
	if update.Version == t.Meta.Version+1 {
		t.Meta.Version = update.Version
	} else {
		return fmt.Errorf("tried to apply update with version %d, task %s has version %d", update.Version, t.UUID, t.Meta.Version)
	}

	if update.EncResult != nil {
		t.EncResult = update.EncResult
	}

	if update.Status != "" && t.Status != update.Status {
		if t.CanTransitionToState(update.Status) {
			if logIt {
				log.LogInfo(fmt.Sprintf("task %s status updated (%s -> %s)", t.UUID, t.Status, update.Status))
			}
			t.Status = update.Status
		} else {
			return fmt.Errorf("task %s tried to transition from %s to %s, throwing update away", t.UUID, t.Status, update.Status)
		}
	}

	if update.RunnerUUID != "" && t.Meta.RunnerUUID != update.RunnerUUID {
		if logIt {
			log.LogInfo(fmt.Sprintf("task %s assigned to runner %s", t.UUID, update.RunnerUUID))
		}
		t.Meta.RunnerUUID = update.RunnerUUID
	}

	if update.RunnerEncTaskKey != nil && t.Meta.RunnerEncTaskKey != update.RunnerEncTaskKey {
		if logIt {
			log.LogInfo(fmt.Sprintf("task %s runner key updated (message encrypted with KID %s)", t.UUID, update.RunnerEncTaskKey.KID))
		}

		t.Meta.RunnerEncTaskKey = update.RunnerEncTaskKey
	}

	if update.RetrySeconds != 0 && t.Meta.RetrySeconds != update.RetrySeconds {
		if logIt {
			log.LogInfo(fmt.Sprintf("task %s set to retry in %d seconds", t.UUID, update.RetrySeconds))
		}
		t.Meta.RetrySeconds = update.RetrySeconds
	}

	return nil
}

// IsPending is if a task is in pending state
func (t *Task) IsPending() bool {
	return t.Status == TaskStatusPending
}

// IsNotStarted is if a task hasn't even tried to run yet (hasn't been assigned a runner)
func (t *Task) IsNotStarted() bool {
	return t.Status == TaskStatusWaiting || t.Status == TaskStatusRetrying
}

// IsRetrying is if a task hasn't even tried to run yet (hasn't been assigned a runner)
func (t *Task) IsRetrying() bool {
	return t.Status == TaskStatusRetrying
}

// IsRunning is if the task is in the process of being run (has been assigned a runner)
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusQueued || t.Status == TaskStatusRunning
}

// IsFinished is if the task has run and has a result of some sort
func (t *Task) IsFinished() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// CanTransitionToState returns true if a task can go from its current state to new
func (t *Task) CanTransitionToState(new string) bool {
	if t.Status == "" {
		return new == TaskStatusWaiting
	}

	if t.Status == TaskStatusPending {
		return new == TaskStatusWaiting
	}

	if t.Status == TaskStatusWaiting {
		return new == TaskStatusQueued || new == TaskStatusRetrying
	}

	if t.Status == TaskStatusQueued {
		return new == TaskStatusRunning || new == TaskStatusFailed || new == TaskStatusRetrying
	}

	if t.Status == TaskStatusRunning {
		return new == TaskStatusCompleted || new == TaskStatusFailed || new == TaskStatusRetrying
	}

	if t.Status == TaskStatusFailed {
		return new == TaskStatusRetrying
	}

	if t.Status == TaskStatusRetrying {
		return new == TaskStatusQueued
	}

	if t.Status == TaskStatusCompleted {
		return false // just want to be real clear that this should never happen
	}

	return false
}
