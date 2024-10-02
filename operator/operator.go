package operator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"go.uber.org/zap"
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
	_, err := aptos.NewClient(op.network)
	if err != nil {
		return fmt.Errorf("failed to create aptos client: %v", err)
	}

	// // looping
	// for {
	// 	// TODO: make a temp variable to track for which latest version to search for events
	// 	event, err := listenForEvent(client, op.avsAddress)
	// 	if err != nil {
	// 		op.logger.Warn("Failed to subscribe to new tasks", zap.Any("err", err))
	// 		time.Sleep(RetryInterval)
	// 		continue
	// 	}
	// 	// TODO: handle task queue full
	// 	agg.TaskQueue <- event // Send event to the task queue
	// 	agg.logger.Info("Queued new task for processing", zap.Any("event", event))
	// 	return nil
	// }
	// TODO
	return nil
}

func (op *Operator) RespondTask(ctx context.Context) error {
	// TODO
	return nil
}
