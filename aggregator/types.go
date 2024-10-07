package aggregator

import (
	"math/big"

	"github.com/aptos-labs/aptos-go-sdk"
	"go.uber.org/zap"
)

type AggregatorConfig struct {
	ServerIpPortAddress string
	AvsAddress          string
	AccountConfig       AccountConfig
}

type AccountConfig struct {
	AccountPath string
	Profile     string
}

type Aggregator struct {
	logger            *zap.Logger
	AvsAddress        string
	AggregatorAccount aptos.Account
	AggregatorConfig  AggregatorConfig
	TaskQueue         chan map[string]interface{}
	Network           aptos.NetworkConfig
}

type SignedTaskResponse struct {
	Pubkey    []byte
	Signature []byte
	Response  big.Int
}
