package operator

import (
	"encoding/hex"
	"fmt"
	"log"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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
	avsAddress := aptos.AccountAddress{}
	if err := avsAddress.ParseStringRelaxed(config.AvsAddress); err != nil {
		panic("Failed to parse avsAddress:" + err.Error())
	}
	registered := IsOperatorRegistered(client, avsAddress, operator_account.Address.String())

	if !registered {
		log.Println("Operator is not registered with A2D Oracle AVS, registering...")

		quorumCount := QuorumCount(client, avsAddress)
		if quorumCount == 0 {
			panic("No quorum found, please initialize quorum first ")
		}

		quorumNumbers := []uint64{quorumCount}

		// Register Operator
		// TODO: minh help
		// ignore error here because panic all the time
		var priv crypto.BlsPrivateKey
		msg := []byte("PubkeyRegistration")
		bcsOperatorAccount, err := bcs.Serialize(&operator_account.Address)
		if err != nil {
			panic("Failed to bsc serialize account" + err.Error())
		}

		msg = append(msg, bcsOperatorAccount...)
		keccakMsg := ethcrypto.Keccak256(msg)
		err = priv.FromBytes(config.BlsPrivateKey)
		if err != nil {
			panic("Failed to create bls priv key" + err.Error())
		}

		signature, err := priv.Sign(keccakMsg)
		if err != nil {
			panic("Failed to create signature" + err.Error())
		}
		pop, err := priv.GenerateBlsPop()
		if err != nil {
			panic("Failed to generate bls proof of possession" + err.Error())
		}
		_ = RegisterOperator(
			client,
			operator_account,
			avsAddress.String(),
			quorumNumbers,
			signature.Auth.Signature().Bytes(),
			signature.PubKey().Bytes(),
			pop.Bytes(),
		)
	}

	// connect to aggregator
	// NewAggregatorRpcClient()
	aggClient, err := NewAggregatorRpcClient(config.AggregatorIpPortAddr)
	// if err != nil {
	// 	return nil, fmt.Errorf("can not create aggregator rpc client: %s", err)
	// }

	// Get OperatorId
	var privKey crypto.BlsPrivateKey
	privKey.FromBytes(config.BlsPrivateKey)
	operatorId := privKey.Inner.PublicKey().Marshal()

	// return Operator
	operator := Operator{
		account:      operator_account,
		operatorId:   operatorId,
		avsAddress:   avsAddress,
		AggRpcClient: *aggClient,
	}
	return &operator, nil
}

func InitQuorum(networkConfig aptos.NetworkConfig, config OperatorConfig, accountConfig AptosAccountConfig) error {

	return nil
}
func QuorumCount(client *aptos.Client, contract aptos.AccountAddress) uint64 {
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
		panic("Could not get quorum count:" + err.Error())
	}
	count := vals[0].(float64)
	return uint64(count)
}

func IsOperatorRegistered(client *aptos.Client, contract aptos.AccountAddress, operator_addr string) bool {

	account := aptos.AccountAddress{}
	err := account.ParseStringRelaxed(operator_addr)
	if err != nil {
		panic("Could not ParseStringRelaxed:" + err.Error())
	}
	operator, err := bcs.Serialize(&account)
	if err != nil {
		panic("Could not serialize operator address:" + err.Error())
	}
	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "registry_coordinator",
		},
		Function: "get_operator_status",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			operator,
		},
	}

	vals, err := client.View(payload)
	if err != nil {
		panic("Could not get operator status:" + err.Error())
	}
	status := vals[0].(float64)
	return status != 0
}

// quorum_numbers: vector<u8>, operator: &signer, params: bls_apk_registry::PubkeyRegistrationParams
func RegisterOperator(
	client *aptos.Client,
	operator_account *aptos.Account,
	contract_addr string,
	quorum_numbers []uint64,
	signature []byte,
	pubkey []byte,
	proofPossession []byte,
) error {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contract_addr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	quorumSerializer := &bcs.Serializer{}
	bcs.SerializeSequence(quorum_numbers, quorumSerializer)

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
	fmt.Println("sig: ", hex.EncodeToString(signature))
	fmt.Println("pubkey: ", hex.EncodeToString(pubkey))
	fmt.Println("proofPossession: ", hex.EncodeToString(proofPossession))
	payload := aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: contract,
			Name:    "registry_coordinator",
		},
		Function: "registor_operator",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			quorumSerializer.ToBytes(), sig, pk, pop,
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
