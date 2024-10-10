package aggregator

import (
	"math/big"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"go.uber.org/zap"
)

type AggregatorConfig struct {
	ServerIpPortAddress string
	AvsAddress          string
	AccountConfig       AccountConfig
}

type AccountConfig struct {
	AccountPath string
	Profile     string
}

type Aggregator struct {
	logger            *zap.Logger
	AvsAddress        string
	AggregatorAccount aptos.Account
	AggregatorConfig  AggregatorConfig
	TaskQueue         chan map[string]interface{}
	Network           aptos.NetworkConfig
}

type SignedTaskResponse struct {
	TaskId    uint64
	Pubkey    []byte
	Signature []byte
	Response  *big.Int
}

type U128Struct struct {
	Value *big.Int `json:"value"`
}

func (u *U128Struct) MarshalBCS(ser *bcs.Serializer) {
	ser.U128(*u.Value)
}

type BytesStruct struct {
	Value []byte
}

func (b *BytesStruct) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(b.Value)
}

type U8Struct struct {
	Value uint8
}

func (u *U8Struct) MarshalBCS(ser *bcs.Serializer) {
	ser.U8(u.Value)
}

type VecAddr struct {
	Value []aptos.AccountAddress
}

func (v *VecAddr) MarshalBCS(ser *bcs.Serializer) {
	bcs.SerializeSequence(v.Value, ser)
}

type Addr struct {
	Value aptos.AccountAddress
}

func (v *Addr) MarshalBCS(ser *bcs.Serializer) {
	v.Value.MarshalBCS(ser)
}
