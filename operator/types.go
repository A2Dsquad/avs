package operator

import (
	"math/big"
	"net/rpc"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"golang.org/x/crypto/ed25519"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
)

type Config struct {
	CmcApi string `json:"cmc_api"`
}

type Operator struct {
	account *aptos.Account
	// TODO: change this to aptos-sdk fork
	operatorId   []byte
	avsAddress   aptos.AccountAddress
	AggRpcClient AggregatorRpcClient
}
type OperatorConfig struct {
	BlsPrivateKey        []byte
	AvsAddress           string
	AggregatorIpPortAddr string
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

type OperatorSetParam struct {
	max_operator_count uint32
}

type MetadataStr struct {
	Inner string
}

type Metadata struct {
	Inner aptos.AccountAddress
}

type U128Struct struct {
	Value *big.Int `json:"value"`
}

func (u *U128Struct) MarshalBCS(ser *bcs.Serializer) {
	ser.U128(*u.Value)
}

type U8Vec struct {
	Value uint8
}

func (u *U8Vec) MarshalBCS(ser *bcs.Serializer) {
	ser.U8(u.Value)
}

type Signature struct {
	bytes []byte
}

type PublicKey struct {
	bytes []byte
}

type Pop struct {
	bytes []byte
}

type AggregatorRpcClient struct {
	rpcClient            *rpc.Client
	aggregatorIpPortAddr string
}
