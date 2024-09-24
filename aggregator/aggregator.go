package aggregator

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewAggregator(aggregatorConfig AggregatorConfig, logger *zap.Logger) (Aggregator, error) {

	aggegator_account, err := SignerFromConfig(aggregatorConfig.accountPath)
	if err != nil {
		return Aggregator{}, errors.Wrap(err, "Failed to create aggregator account")
	}
	agg := Aggregator{
		logger:            logger,
		AggregatorAccount: *aggegator_account,
		AggregatorConfig:  aggregatorConfig,
	}
	return agg, nil
}
