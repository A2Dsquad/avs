package aggregator

import (
	"context"
	"fmt"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

func (agg *Aggregator) DoChore(ctx context.Context) error {
	client, err := aptos.NewClient(agg.Network)
	if err != nil {
		return fmt.Errorf("failed to create aptos client: %v", err)
	}

	avsAddress := aptos.AccountAddress{}
	err = avsAddress.ParseStringRelaxed(agg.AvsAddress)
	if err != nil {
		return fmt.Errorf("failed to parse avs address: %v", err)
	}
	// Get quorum count
	quorumCount, err := QuorumCount(client, avsAddress)
	if err != nil {
		return fmt.Errorf("failed to get quorum count: %v", err)
	}

	var operatorsPerQuorum [][]interface{}

	if quorumCount != 0 {
		for i := 1; i <= int(quorumCount); i++ {
			operatorList, err := GetOperatorListAtTimestamp(client, avsAddress, uint8(i), uint64(time.Now().Unix()))
			fmt.Println("test :", err)
			operatorsPerQuorum = append(operatorsPerQuorum, operatorList)
		}
	}

	fmt.Println("operatorsPerQuorum : ", operatorsPerQuorum)

	// Get list quorum
	// Get list operator for each quorum
	// Update quorums

	return nil
}

func QuorumCount(client *aptos.Client, contract aptos.AccountAddress) (uint8, error) {
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "registry_coordinator",
		},
		Function: "quorum_count",
		ArgTypes: []aptos.TypeTag{},
		Args:     [][]byte{},
	}

	vals, err := client.View(payload)
	if err != nil {
		return 0, fmt.Errorf("No quorum found")
	}
	count := vals[0].(float64)
	return uint8(count), nil
}

func GetOperatorListAtTimestamp(client *aptos.Client, contract aptos.AccountAddress, quorum uint8, timestamp uint64) ([]interface{}, error) {
	quorumBcs, err := bcs.SerializeU8(quorum)
	if err != nil {
		return nil, err
	}
	timestampBcs, err := bcs.SerializeU64(timestamp)
	if err != nil {
		return nil, err
	}
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "index_registry",
		},
		Function: "get_operator_list_at_timestamp",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			quorumBcs, timestampBcs,
		},
	}

	vals, err := client.View(payload)
	if err != nil {
		return nil, err
	}
	operatorList := vals[0].([]interface{})
	return operatorList, nil
}
