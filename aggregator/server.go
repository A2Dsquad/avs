package aggregator

import (
	"fmt"
	"math/big"
	"net/http"
	"net/rpc"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
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

// public entry fun respond_to_task(
// 	aggregator: &signer,
// 	task_id: u64,
// 	responses: vector<u128>,
// 	signer_pubkeys: vector<vector<u8>>,
// 	signer_sigs: vector<vector<u8>>,
// 	quorum_aggr_pks: vector<vector<u8>>,
// ) acquires ServiceManagerStore {

func RespondToAvs(
	client *aptos.Client,
	operatorAccount *aptos.Account,
	contractAddr string,
	taskId uint64,
	signature [][]byte,
	pubkey [][]byte,
	responses []big.Int,
	quorum_aggr_pks 
) error {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contractAddr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	quorumSerializer := &bcs.Serializer{}
	bcs.SerializeSequence([]U8Struct{
		{
			Value: quorumNumbers,
		},
	}, quorumSerializer)

	sig, err := bcs.SerializeBytes(signature)
	if err != nil {
		panic("Failed to bcs serialize signature:" + err.Error())
	}
	pk, err := bcs.SerializeBytes(pubkey)
	if err != nil {
		panic("Failed to bcs serialize pubkey:" + err.Error())
	}
	pop, err := bcs.SerializeBytes(proofPossession)
	if err != nil {
		panic("Failed to bcs serialize proof of possession:" + err.Error())
	}
	payload := aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "service_manager",
		},
		Function: "respond_to_task",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			quorumSerializer.ToBytes(), sig, pk, pop,
		},
	}
	// Build transaction
	rawTxn, err := client.BuildTransaction(operatorAccount.AccountAddress(),
		aptos.TransactionPayload{Payload: &payload})
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(operatorAccount)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}
	fmt.Printf("Submit register operator for %s\n", operatorAccount.AccountAddress())

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
