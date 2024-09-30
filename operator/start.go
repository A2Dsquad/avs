package operator

import (
	"fmt"
	"log"
	"math/big"

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
	operatorAccount, err := SignerFromConfig(accountConfig.configPath, accountConfig.profile)
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
	registered := IsOperatorRegistered(client, avsAddress, operatorAccount.Address.String())

	if !registered {
		log.Println("Operator is not registered with A2D Oracle AVS, registering...")

		quorumCount := QuorumCount(client, avsAddress)
		if quorumCount == 0 {
			panic("No quorum found, please initialize quorum first ")
		}

		quorumNumbers := quorumCount

		// Register Operator
		// ignore error here because panic all the time
		var priv crypto.BlsPrivateKey
		msg := []byte("PubkeyRegistration")
		bcsOperatorAccount, err := bcs.Serialize(&operatorAccount.Address)
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
			operatorAccount,
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
		account:      operatorAccount,
		operatorId:   operatorId,
		avsAddress:   avsAddress,
		AggRpcClient: *aggClient,
	}
	return &operator, nil
}

func InitQuorum(
	networkConfig aptos.NetworkConfig,
	config OperatorConfig,
	accountConfig AptosAccountConfig,
	maxOperatorCount uint32,
	minimumStake big.Int,
) error {
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	accAddress := aptos.AccountAddress{}
	err = accAddress.ParseStringRelaxed("0x603053371d0eec6befaf41489f506b7b3e8e31dbca3d9b9c5cb92bb308dc2eec")
	if err != nil {
		panic("Failed to parse account address " + err.Error())
	}

	operatorAccount, err := SignerFromConfig(accountConfig.configPath, accountConfig.profile)
	if err != nil {
		panic("Failed to create operator account:" + err.Error())
	}

	maxOperatorCountBz, err := bcs.SerializeU32(maxOperatorCount)
	if err != nil {
		return fmt.Errorf("failed to serialize maxOperatorCount: %s", err)
	}

	minimumStakeBz, err := bcs.SerializeU128(minimumStake)
	if err != nil {
		return fmt.Errorf("failed to serialize minimumStake: %s", err)
	}

	// "0xcc28657ec961d4a93d2b7a89853fb73912a45dd582cff9fa43ebe7fd7ec26799"
	// faMetadata := GetMetadata(client)
	// var test []Metadata

	// faClient := FAMetdataClient(client)
	// faStoreAddr, err := faClient.PrimaryStoreAddress(&accAddress)
	// if err != nil {
	// 	panic("Failed to ")
	// }
	// fmt.Println("faStoreAddr: ", faStoreAddr.String())
	// hex, _:= hex.DecodeString("0x0572757065650100000000000000004038393534453933384132434137314536433445313139434230333341363036453341333537424245353843354430304235453132354236383238423745424331e7011f8b08000000000002ff3d8ecd6ec4200c84ef3cc58a7b13427ea9d4432f7d8928aa0c7636d136100149fbf885ed764ff68cbeb167dcc1dce04a13b3b0d1e5edc2fdb1137176920fabb3d9a90a5108cee0888bf32139e3c4d808889e42a030b17be4331b19173faa1f8cac6aa5846986bac6b9afda6648c350a3b06d4cdf2aec7487aa9648526ad960db894ee81e6609c0d379a49d2c92352b85e27d8f2e7cf8d4f0dbf9dbc4ae6bcc9f9618f7f05a96492e872e8cdb4ac8e4cb17e8f0588df3542480334f670e6db05a4b498743e37a6ffc476eeea472fe7ff2883f3567bf9aa419822b010000010572757065650000000300000000000000000000000000000000000000000000000000000000000000010e4170746f734672616d65776f726b00000000000000000000000000000000000000000000000000000000000000010b4170746f735374646c696200000000000000000000000000000000000000000000000000000000000000010a4d6f76655374646c696200")
	// fmt.Println("hex: ", string(hex))
	// bcs.Deserialize()
	// addr := aptos.AccountAddress{}
	// addr.ParseStringRelaxed("0xcc28657ec961d4a93d2b7a89853fb73912a45dd582cff9fa43ebe7fd7ec26799")
	// metadataBz, err := json.Marshal(Metadata{
	// 	Inner: addr,
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal metadata: %s", err)
	// }
	// metadataHex := hex.EncodeToString(metadataBz)
	// fmt.Println("bz: ", metadataHex)

	metadataAddr := GetMetadata(client).Inner

	strategiesSerializer := &bcs.Serializer{}
	bcs.SerializeSequence([]aptos.AccountAddress{metadataAddr}, strategiesSerializer)

	multiplier := new(big.Int)
	multiplier.SetString("10000000", 10)

	multipliersSerializer := &bcs.Serializer{}
	bcs.SerializeSequence([]U128Struct{{
		Value: multiplier,
	}}, multipliersSerializer)

	// Get operator Status
	avsAddress := aptos.AccountAddress{}
	if err := avsAddress.ParseStringRelaxed(config.AvsAddress); err != nil {
		panic("Failed to parse avsAddress:" + err.Error())
	}

	payload := aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: avsAddress,
			Name:    "registry_coordinator",
		},
		Function: "create_quorum",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			maxOperatorCountBz, minimumStakeBz, strategiesSerializer.ToBytes(), multipliersSerializer.ToBytes(),
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
		panic("Failed to create quorum")
	}
	return nil
}
func QuorumCount(client *aptos.Client, contract aptos.AccountAddress) uint8 {
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
	return uint8(count)
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
	operatorAccount *aptos.Account,
	contractAddr string,
	quorumNumbers uint8,
	signature []byte,
	pubkey []byte,
	proofPossession []byte,
) error {
	contract := aptos.AccountAddress{}
	err := contract.ParseStringRelaxed(contractAddr)
	if err != nil {
		panic("Failed to parse address:" + err.Error())
	}
	quorumSerializer := &bcs.Serializer{}
	bcs.SerializeSequence([]U8Vec{
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
			Name:    "registry_coordinator",
		},
		Function: "registor_operator",
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
