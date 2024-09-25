package operator

import (
	"fmt"
	"log"

	// For now using eigentypes
	eigentypes "github.com/Layr-Labs/eigensdk-go/types"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

type AptosAccountConfig struct {
	configPath string
	profile    string
}

func AptosClient(networkConfig aptos.NetworkConfig) *aptos.Client {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}
	return client
}

func NewOperator(networkConfig aptos.NetworkConfig, config OperatorConfig, accountConfig AptosAccountConfig) (*Operator, error) {
	operator_account, err := SignerFromConfig(accountConfig.configPath, accountConfig.profile)
	if err != nil {
		panic("Failed to create operator account:" + err.Error())
	}
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Get operator Status
	registered := IsOperatorRegistered(client, config.AvsAddress, config.OperatorAddress.String())

	if !registered {
		log.Println("Operator is not registered with A2D Oracle AVS, registering...")

		quorumNumbers := []byte{0}

		// Register Operator
		// TODO: minh help
		// ignore error here because panic all the time
		_ = RegisterOperator(client, operator_account, config.AvsAddress.String(), quorumNumbers, PubkeyRegistrationParams{})
	}

	// connect to aggregator
	// NewAggregatorRpcClient()
	aggClient, err := NewAggregatorRpcClient(config.aggregatorIpPortAddr)
	if err != nil {
		return nil, fmt.Errorf("can not create aggregator rpc client: %s", err)
	}

	// Get OperatorId
	operatorId := eigentypes.OperatorIdFromKeyPair(config.BlsKeyPair)

	// return Operator
	operator := Operator{
		account:      operator_account,
		operatorId:   operatorId,
		avsAddress:   config.AvsAddress,
		AggRpcClient: *aggClient,
	}
	return &operator, nil
}

func IsOperatorRegistered(client *aptos.Client, contract aptos.AccountAddress, operator_addr string) bool {
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "registry_coordinator",
		},
		Function: "get_operator_status",
		ArgTypes: []aptos.TypeTag{},
		Args:     [][]byte{[]byte(operator_addr)},
	}

	vals, err := client.View(payload)
	if err != nil {
		panic("Could not get operator status:" + err.Error())
	}
	status := vals[0].(uint8)
	return status == 0
}

// quorum_numbers: vector<u8>, operator: &signer, params: bls_apk_registry::PubkeyRegistrationParams
func RegisterOperator(
	client *aptos.Client,
	operator_account *aptos.Account,
	contract_addr string,
	quorum_numbers []byte,
	bls_register_params PubkeyRegistrationParams,
) error {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contract_addr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	quorum, err := bcs.SerializeBytes(quorum_numbers)
	if err != nil {
		panic("Failed to bcs serialize quorum:" + err.Error())
	}
	params, err := bcs.Serialize(&bls_register_params)
	if err != nil {
		panic("Failed to bcs serialize bls_register_params:" + err.Error())
	}
	payload := aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "registry_coordinator",
		},
		Function: "registor_operator",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			quorum, operator_account.Signer.AuthKey().Bytes(), params,
		},
	}
	// Build transaction
	rawTxn, err := client.BuildTransaction(operator_account.AccountAddress(),
		aptos.TransactionPayload{Payload: &payload})
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(operator_account)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}
	fmt.Printf("Submit register operator for %s\n", operator_account.AccountAddress())

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// Wait for the transaction
	fmt.Printf("And we wait for the transaction %s to complete...\n", txnHash)
	userTxn, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	fmt.Printf("The transaction completed with hash: %s and version %d\n", userTxn.Hash, userTxn.Version)
	if !userTxn.Success {
		// TODO: log something more
		panic("Failed to register operator")
	}
	return nil
}
