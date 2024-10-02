package operator

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	// TODO
	return nil
}

func (op *Operator) RespondTask(ctx context.Context) error {
	// TODO
	return nil
} 
