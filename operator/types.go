package operator

import (
	"net/rpc"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"golang.org/x/crypto/ed25519"

	eigentypes "github.com/Layr-Labs/eigensdk-go/types"
)

type Config struct {
	CmcApi string `json:"cmc_api"`
}

type Operator struct {
	account *aptos.Account
	// TODO: change this to aptos-sdk fork
	operatorId   eigentypes.Bytes32
	avsAddress   aptos.AccountAddress
	AggRpcClient AggregatorRpcClient
}
type OperatorConfig struct {
	BlsKeyPair           *bls.KeyPair
	OperatorAddress      aptos.AccountAddress
	AvsAddress           aptos.AccountAddress
	aggregatorIpPortAddr string
	// OperatorId           eigentypes.OperatorId
}

// type OperatorConfig struct {
// 	// TODO: EcdsaConfig

// }

type BlsConfig struct {
	KeyPair *bls.KeyPair
}

type AlternativeSigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

type PubkeyRegistrationParams struct {
	signature Signature
	pubkey_g1 PublicKeyWithPoP
	pubkey_g2 PublicKeyWithPoP
}

type Signature struct {
	bytes []byte
}

type PublicKeyWithPoP struct {
	bytes []byte
}

// Implement MarshalBCS
func (params *PubkeyRegistrationParams) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(params.signature.bytes)
	ser.WriteBytes(params.pubkey_g1.bytes)
	ser.WriteBytes(params.pubkey_g2.bytes)
}

type AggregatorRpcClient struct {
	rpcClient            *rpc.Client
	aggregatorIpPortAddr string
}

type SignedTaskResponse struct {
	BlsSignature bls.Signature
	OperatorId   eigentypes.OperatorId
}
