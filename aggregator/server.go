package aggregator

import (
	"fmt"
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
	client, err := aptos.NewClient(agg.Network)
	if err != nil {
		return fmt.Errorf("failed to create aptos client: %v", err)
	}

	// TODO: aggregate
	err = RespondToAvs(client, &agg.AggregatorAccount, agg.AvsAddress, signedTaskResponse.TaskId,
		[]BytesStruct{{
			Value: signedTaskResponse.Signature,
		}},
		[]BytesStruct{{
			Value: signedTaskResponse.Pubkey,
		}},
		[]U128Struct{{
			Value: signedTaskResponse.Response,
		}},
	)

	if err != nil {
		return fmt.Errorf("failed to respond task: %v", err)
	}
	// build the response
	// query if quorum has sastified
	// create tx
	return nil
}

// aggregator: &signer,
// task_id: u64,
// responses: vector<u128>,
// signer_pubkeys: vector<vector<u8>>,
// signer_sigs: vector<vector<u8>>,

func RespondToAvs(
	client *aptos.Client,
	aggregatorAccount *aptos.Account,
	contractAddr string,
	taskId uint64,
	signature []BytesStruct,
	pubkey []BytesStruct,
	responses []U128Struct,
) error {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contractAddr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	taskIdBcs, err := bcs.SerializeU64(taskId)
	if err != nil {
		panic("Failed to bcs serialize task id:" + err.Error())
	}

	sigSerializer := bcs.Serializer{}
	bcs.SerializeSequence(signature, &sigSerializer)

	pubkeySerializer := bcs.Serializer{}
	bcs.SerializeSequence(pubkey, &pubkeySerializer)

	responseSerializer := bcs.Serializer{}
	bcs.SerializeSequence(responses, &responseSerializer)
	payload := aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "service_manager",
		},
		Function: "respond_to_task",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			taskIdBcs, responseSerializer.ToBytes(), pubkeySerializer.ToBytes(), sigSerializer.ToBytes(),
		},
	}
	fmt.Println("aggregatorAccount.AccountAddress() :", aggregatorAccount.AccountAddress())
	// Build transaction
	rawTxn, err := client.BuildTransaction(aggregatorAccount.AccountAddress(),
		aptos.TransactionPayload{Payload: &payload})
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(aggregatorAccount)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}
	fmt.Printf("Submit register operator for %s\n", aggregatorAccount.AccountAddress())

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
		panic("Failed to respond to avs")
	}
	return nil
}
