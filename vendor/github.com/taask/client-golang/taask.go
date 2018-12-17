package taask

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cohix/simplcrypto"

	log "github.com/cohix/simplog"
	"github.com/pkg/errors"
	"github.com/taask/taask-server/model"
	"github.com/taask/taask-server/service"
	"google.golang.org/grpc"
)

// Client describes a taask client
type Client struct {
	client   service.TaskServiceClient
	taskKeys map[string]*simplcrypto.KeyPair
	keyLock  *sync.Mutex
}

// NewClient creates a Client
func NewClient(addr, port string) (*Client, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", addr, port), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to Dial")
	}

	client := &Client{
		taskKeys: make(map[string]*simplcrypto.KeyPair),
		keyLock:  &sync.Mutex{},
	}

	client.client = service.NewTaskServiceClient(conn)

	return client, nil
}

// SendTask sends a task to be run
func (c *Client) SendTask(task *model.Task) (string, error) {
	if task.Meta == nil {
		task.Meta = &model.TaskMeta{}
	}

	taskKeyPair, err := simplcrypto.GenerateNewKeyPair()
	if err != nil {
		return "", errors.Wrap(err, "failed to GenerateNewKeyPair")
	}

	task.Meta.ResultPubKey = taskKeyPair.SerializablePubKey()

	resp, err := c.client.Queue(context.Background(), task)
	if err != nil {
		return "", errors.Wrap(err, "failed to Queue")
	}

	c.keyLock.Lock()
	c.taskKeys[resp.UUID] = taskKeyPair
	c.keyLock.Unlock()

	return resp.UUID, nil
}

// GetTaskResult gets a task's result
func (c *Client) GetTaskResult(uuid string) ([]byte, error) {
	c.keyLock.Lock()
	taskKeyPair, ok := c.taskKeys[uuid]
	if !ok {
		c.keyLock.Unlock()
		return nil, errors.New(fmt.Sprintf("unable to find task %s key", uuid))
	}
	c.keyLock.Unlock()

	stream, err := c.client.CheckTask(context.Background(), &model.CheckTaskRequest{UUID: uuid})
	if err != nil {
		return nil, errors.Wrap(err, "failed to CheckTask")
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, errors.Wrap(err, "failed to Recv")
		}

		log.LogInfo(fmt.Sprintf("task %s status %s", uuid, resp.Status))

		if resp.Status == model.TaskStatusCompleted {
			result, err := decryptResult(taskKeyPair, resp.Result.EncResultSymKey, resp.Result.EncResult)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decryptResult for complete task")
			}

			return result, nil
		} else if resp.Status == model.TaskStatusFailed {
			// do nothing for now
		}

		<-time.After(time.Second)
	}
}

func decryptResult(taskKeyPair *simplcrypto.KeyPair, encResultKey *simplcrypto.Message, encResult *simplcrypto.Message) ([]byte, error) {
	decResultKeyJSON, err := taskKeyPair.Decrypt(encResultKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Decrypt result key")
	}

	resultKey, err := simplcrypto.SymKeyFromJSON(decResultKeyJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to SymKeyFromJSON")
	}

	decResult, err := resultKey.Decrypt(encResult)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Decrypt result")
	}

	return decResult, nil
}
