package aggregator

import (
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"go.uber.org/zap"

	eigentypes "github.com/Layr-Labs/eigensdk-go/types"
)

type AggregatorConfig struct {
	ServerIpPortAddress string
	avsAddress          string
	accountConfig       AccountConfig
}

type AccountConfig struct {
	accountPath string
	profile     string
}

type Aggregator struct {
	logger            *zap.Logger
	AvsAddress        string
	AggregatorAccount aptos.Account
	AggregatorConfig  AggregatorConfig
	TaskQueue         chan api.EventV2
}

type SignedTaskResponse struct {
	BlsSignature crypto.BlsSignature
	OperatorId   eigentypes.OperatorId
}
