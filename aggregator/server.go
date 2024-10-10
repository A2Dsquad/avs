package aggregator

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"go.uber.org/zap"
)

const (
	THRESHOLD_DENOMINATOR       uint64 = 100
	QUORUM_THRESHOLD_PERCENTAGE uint64 = 67
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

	// GetMsgHashes
	msgHashes, err := GetMsgHashes(client, agg.AvsAddress, signedTaskResponse.TaskId,
		[]U128Struct{{
			Value: signedTaskResponse.Response,
		}},
		[]BytesStruct{{
			Value: signedTaskResponse.Pubkey,
		}},
	)
	if err != nil {
		return fmt.Errorf("lmaaaooo: %v", err)
	}

	hexStr, ok := msgHashes[0].(string)
	if !ok {
		return fmt.Errorf("data is not a string")
	}
	trimmedHexStr := strings.TrimPrefix(hexStr, "0x")
	bytesMsgHash, err := hex.DecodeString(trimmedHexStr)
	if err != nil {
		return fmt.Errorf("can't decode string: %v", err)
	}

	signedStake, totalStake, err := CheckSignatures(client, agg.AvsAddress, 1, uint64(1728544933),
		[]BytesStruct{{
			Value: bytesMsgHash,
		}},
		[]BytesStruct{{
			Value: signedTaskResponse.Pubkey,
		}},
		[]BytesStruct{{
			Value: signedTaskResponse.Signature,
		}},
	)
	if err != nil {
		return fmt.Errorf("can't check signature: %v", err)
	}

	// (signed_stake * THRESHOLD_DENOMINATOR) >= (total_stake * QUORUM_THRESHOLD_PERCENTAGE)
	if signedStake*THRESHOLD_DENOMINATOR >= totalStake*QUORUM_THRESHOLD_PERCENTAGE {
		agg.logger.Info("Quorum for task has reached. Responding...", zap.Any("task_id", signedTaskResponse.TaskId))
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
	} else {
		agg.logger.Info("Quorum for task has not reached. Waiting for other operators", zap.Any("task_id", signedTaskResponse.TaskId))
		// TODO: save pub and sig for this task for later aggregate
	}
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

// quorum_numbers: vector<u8>,
// reference_timestamp: u64,
// msg_hashes: vector<vector<u8>>,
// signer_pubkeys: vector<vector<u8>>,
// signer_sigs: vector<vector<u8>>,
func CheckSignatures(
	client *aptos.Client,
	contractAddr string,
	quorumNumbers uint8,
	referenceTimestamp uint64,
	msgHashes []BytesStruct,
	pubkey []BytesStruct,
	signature []BytesStruct,
) (uint64, uint64, error) {
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

	timestampBcs, err := bcs.SerializeU64(referenceTimestamp)
	if err != nil {
		panic("Failed to SerializeU64:" + err.Error())
	}

	sigSerializer := bcs.Serializer{}
	bcs.SerializeSequence(signature, &sigSerializer)

	pubkeySerializer := bcs.Serializer{}
	bcs.SerializeSequence(pubkey, &pubkeySerializer)

	msgHashesSerializer := bcs.Serializer{}
	bcs.SerializeSequence(msgHashes, &msgHashesSerializer)

	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "bls_sig_checker",
		},
		Function: "check_signatures",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			quorumSerializer.ToBytes(),
			timestampBcs,
			msgHashesSerializer.ToBytes(),
			pubkeySerializer.ToBytes(),
			sigSerializer.ToBytes(),
		},
	}

	vals, err := client.View(payload)
	if err != nil {
		return 0, 0, err
	}
	signedStakeStr := vals[0].([]interface{})[0].(string)
	signedStake, err := strconv.ParseUint(signedStakeStr, 10, 64) // base 10, 64-bit size
	if err != nil {
		return 0, 0, fmt.Errorf("error converting string to uint64: %v", err)
	}
	totalStakeStr := vals[1].([]interface{})[0].(string)
	totalStake, err := strconv.ParseUint(totalStakeStr, 10, 64) // base 10, 64-bit size
	if err != nil {
		return 0, 0, fmt.Errorf("error converting string to uint64: %v", err)
	}
	return signedStake, totalStake, nil
}

func GetMsgHashes(
	client *aptos.Client,
	contractAddr string,
	taskId uint64,
	responses []U128Struct,
	pubkey []BytesStruct,
) ([]interface{}, error) {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contractAddr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}

	taskIdBcs, err := bcs.SerializeU64(taskId)
	if err != nil {
		panic("Failed to bcs serialize task id:" + err.Error())
	}

	pubkeySerializer := bcs.Serializer{}
	bcs.SerializeSequence(pubkey, &pubkeySerializer)

	responseSerializer := bcs.Serializer{}
	bcs.SerializeSequence(responses, &responseSerializer)

	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "service_manager",
		},
		Function: "get_msg_hashes",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			taskIdBcs,
			responseSerializer.ToBytes(),
			pubkeySerializer.ToBytes(),
		},
	}

	vals, err := client.View(payload)
	if err != nil {
		return nil, err
	}

	msgHashes := vals[0].([]interface{})
	return msgHashes, nil
}
