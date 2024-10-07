
package aggregator

import (
	"fmt"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"go.uber.org/zap"
)

const (
	MaxRetries                        = 100
	RetryInterval                     = 1 * time.Second
	BlockInterval              uint64 = 1000
	PollLatestBatchInterval           = 5 * time.Second
	RemoveBatchFromSetInterval        = 5 * time.Minute
)

func (agg *Aggregator) SubscribeToNewTasks(network aptos.NetworkConfig) error {
	client, err := aptos.NewClient(network)
	if err != nil {
		return fmt.Errorf("failed to create aptos client: %v", err)
	}

	// looping
	for {
		// TODO: make a temp variable to track for which latest version to search for events
		event, err := listenForEvent(client, agg.AvsAddress)
		if err != nil {
			agg.logger.Warn("Failed to subscribe to new tasks", zap.Any("err", err))
			time.Sleep(RetryInterval)
			continue
		}

		// TODO: handle task queue full
		agg.TaskQueue <- event // Send event to the task queue
		agg.logger.Info("Queued new task for processing", zap.Any("event", event))
		return nil
	}
}

func listenForEvent(client *aptos.Client, avsAddr string) (api.EventV2, error) {
	// client.AccountsEvents()
	return api.EventV2{}, nil
}
