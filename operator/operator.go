package operator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"go.uber.org/zap"
)

const (
	MaxRetries                        = 100
	RetryInterval                     = 1 * time.Second
	BlockInterval              uint64 = 1000
	PollLatestBatchInterval           = 5 * time.Second
	RemoveBatchFromSetInterval        = 5 * time.Minute
)

func (op *Operator) Start(ctx context.Context) error {
	op.logger.Info("Starting operator...")

	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	// Fetching tasks
	go func() {
		op.logger.Info("Fetching tasks process started...")
		err := op.FetchTasks(ctx)
		if err != nil {
			op.logger.Fatal("Error listening for tasks", zap.Any("err", err))
		}
	}()

	go func() {
		op.logger.Info("Respond tasks process started...")
		err := op.RespondTask(ctx)
		if err != nil {
			op.logger.Fatal("Error listening for tasks", zap.Any("err", err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal to shutdown
	sig := <-sigChan
	op.logger.Info("Received signal, shutting down...", zap.Any("signal", sig))

	cancel()
	return nil
}

func (op *Operator) FetchTasks(ctx context.Context) error {
	client, err := aptos.NewClient(op.network)
	if err != nil {
		return fmt.Errorf("failed to create aptos client: %v", err)
	}

	var taskCount uint64
	// looping
	for {
		previousTaskCount := taskCount
		newTaskCount, err := LatestTaskCount(client, op.avsAddress)
		if err != nil {
			op.logger.Warn("Failed to subscribe to new tasks", zap.Any("err", err))
			time.Sleep(RetryInterval)
			continue
		}
		taskCount = newTaskCount

		if taskCount > previousTaskCount {
			err := op.QueueTask(ctx, client, previousTaskCount, taskCount)
			if err != nil {
				return fmt.Errorf("error queuing task: %v", err)
			}
		}

		// TODO: handle task queue full
		// agg.TaskQueue <- event // Send event to the task queue
		// agg.logger.Info("Queued new task for processing", zap.Any("event", event))
	}
	// TODO
	return nil
}

func (op *Operator) RespondTask(ctx context.Context) error {
	// TODO
	return nil
}

func (op *Operator) QueueTask(ctx context.Context, client *aptos.Client, start uint64, end uint64) error {
	for i := start + 1; i <= end; i++ {
		task, err := LoadTaskById(client, op.avsAddress, i)
		if err != nil {
			return fmt.Errorf("error loading task: %v", err)
		}
		op.logger.Info("Loaded new task with id: %d", zap.Any("task id", i))
		op.TaskQueue <- task
		op.logger.Info("Queued new task with id: %d", zap.Any("task id", i))
	}

	return nil
}

func LoadTaskById(client *aptos.Client, contract aptos.AccountAddress, taskId uint64) (map[string]interface{}, error) {
	taskIdBcs, err := bcs.SerializeU64(taskId)
	if err != nil {
		return nil, fmt.Errorf("can not SerializeU64: %v", err)
	}
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "service_manager",
		},
		Function: "task_by_id",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			taskIdBcs,
		},
	}
	vals, err := client.View(payload)
	if err != nil {
		return nil, fmt.Errorf("can not get task count: %v", err)
	}
	task := vals[0].(map[string]interface{})
	return task, nil
}

func LatestTaskCount(client *aptos.Client, contract aptos.AccountAddress) (uint64, error) {
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "service_manager",
		},
		Function: "task_count",
		ArgTypes: []aptos.TypeTag{},
		Args:     [][]byte{},
	}

	vals, err := client.View(payload)
	if err != nil {
		return 0, fmt.Errorf("can not get task count: %v", err)
	}
	countStr := vals[0].(string)

	count, err := strconv.ParseUint(countStr, 10, 64) // base 10 and 64-bit size
	if err != nil {
		return 0, fmt.Errorf("error parsing task count: %s", err)
	}
	return uint64(count), nil
}
