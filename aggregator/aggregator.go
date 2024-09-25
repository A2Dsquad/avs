package aggregator

import (
	"context"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	
)

const taskQueueSize = 100

func NewAggregator(aggregatorConfig AggregatorConfig, logger *zap.Logger) (Aggregator, error) {
	aggegator_account, err := SignerFromConfig(aggregatorConfig.accountConfig.accountPath, aggregatorConfig.accountConfig.profile)
	if err != nil {
		return Aggregator{}, errors.Wrap(err, "Failed to create aggregator account")
	}

	agg := Aggregator{
		logger:            logger,
		AvsAddress:        aggregatorConfig.avsAddress,
		AggregatorAccount: *aggegator_account,
		AggregatorConfig:  aggregatorConfig,
		TaskQueue:         make(chan api.EventV2, taskQueueSize),
	}
	return agg, nil
}

func (agg *Aggregator) Start(ctx context.Context) error {
	agg.logger.Info("Starting aggregator...")
	go func() {
		err := agg.ServeOperators()
		if err != nil {
			agg.logger.Fatal("Error listening for tasks", zap.Any("err", err))
		}
	}()

	return nil
}

