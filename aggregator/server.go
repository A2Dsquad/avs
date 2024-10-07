package aggregator

import (
	"fmt"
	"net/http"
	"net/rpc"

	"go.uber.org/zap"
)

func (agg *Aggregator) ServeOperators() error {
	// Registers a new RPC server
	err := rpc.Register(agg)
	if err != nil {
		return err
	}

	// Registers an HTTP handler for RPC messages
	rpc.HandleHTTP()

	agg.logger.Info("Starting RPC server on address:", zap.String("address", agg.AggregatorConfig.ServerIpPortAddress))

	err = http.ListenAndServe(agg.AggregatorConfig.ServerIpPortAddress, nil)
	if err != nil {
		return err
	}

	return nil
}

// Define the RespondTask method for handling incoming RPC calls
func (agg *Aggregator) RespondTask(signedTaskResponse SignedTaskResponse, reply *uint8) error {
	agg.logger.Info("Received signed task response", zap.Any("response", signedTaskResponse))

	// Process the signed task response
	if err := agg.processTaskResponse(signedTaskResponse); err != nil {
		agg.logger.Error("Failed to process signed task response", zap.Error(err))
		return fmt.Errorf("failed to process task response: %v", err)
	}

	// Set reply to indicate success (e.g., 0 = success)
	*reply = 0
	agg.logger.Info("Successfully processed signed task response")
	return nil
}

func (agg *Aggregator) processTaskResponse(signedTaskResponse SignedTaskResponse) error {

	// build the response
	// query if quorum has sastified
	// create tx
	return nil
}