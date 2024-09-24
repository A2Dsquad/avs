package aggregator

import (
	"net/http"
	"net/rpc"

	"github.com/aptos-labs/aptos-go-sdk"
	"go.uber.org/zap"
)

type AggregatorConfig struct {
	ServerIpPortAddress string
	avsAddress          string
	accountPath         string
}

type Aggregator struct {
	logger            *zap.Logger
	AggregatorAccount aptos.Account
	AggregatorConfig  AggregatorConfig
}

func (agg *Aggregator) ServeOperators() error {
	// Registers a new RPC server
	err := rpc.Register(agg)
	if err != nil {
		return err
	}

	// Registers an HTTP handler for RPC messages
	rpc.HandleHTTP()

	// Start listening for requests on aggregator address
	// ServeOperators accepts incoming HTTP connections on the listener, creating
	// a new service goroutine for each. The service goroutines read requests
	// and then call handler to reply to them
	agg.logger.Info("Starting RPC server on address:", zap.String("address", agg.AggregatorConfig.ServerIpPortAddress))

	err = http.ListenAndServe(agg.AggregatorConfig.ServerIpPortAddress, nil)
	if err != nil {
		return err
	}

	return nil
}
