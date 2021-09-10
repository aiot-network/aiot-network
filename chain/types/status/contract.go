package status

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/types/status/exchange"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/rlp"
)

type ContractType uint
type FunctionType uint

const (
	Exchange_ ContractType = 0
	Pair_                  = 1
	Token_                 = 2
)

const (
	Exchange_Init     FunctionType = 000000
	Exchange_SetAdmin              = 000001
	Exchange_SetFeeTo              = 000002
	Exchange_ExactIn               = 000003
	Exchange_ExactOut              = 000004

	Pair_AddLiquidity    = 100000
	Pair_RemoveLiquidity = 100001

	Token_Create = 200000
)

type Contract struct {
	Address    arry.Address
	CreateHash arry.Hash
	Type       ContractType
	Body       IContractBody
}

func (c *Contract) Bytes() []byte {
	rlpC := &RlpContract{
		Address:    c.Address,
		CreateHash: c.CreateHash,
		Type:       c.Type,
		Body:       c.Body.Bytes(),
	}
	bytes, _ := rlp.EncodeToBytes(rlpC)
	return bytes
}

func (c *Contract) Verify(function FunctionType, sender arry.Address) error {
	ex, _ := c.Body.(*exchange.Exchange)
	switch function {
	case Exchange_Init:
		return fmt.Errorf("exchange %s already exist", c.Address.String())
	case Exchange_SetAdmin:
		return ex.VerifySetter(sender)
	case Exchange_SetFeeTo:
		return ex.VerifySetter(sender)
	}

	return nil
}

type RlpContract struct {
	Address    arry.Address
	CreateHash arry.Hash
	Type       ContractType
	Body       []byte
}

type IContractBody interface {
	Bytes() []byte
}

func DecodeContract(bytes []byte) (*Contract, error) {
	var rlpContract *RlpContract
	if err := rlp.DecodeBytes(bytes, &rlpContract); err != nil {
		return nil, err
	}
	var contract = &Contract{
		Address:    rlpContract.Address,
		CreateHash: rlpContract.CreateHash,
		Type:       rlpContract.Type,
		Body:       nil,
	}
	switch rlpContract.Type {
	case Exchange_:
		ex, err := exchange.DecodeToExchange(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = ex
		return contract, err
	case Pair_:
		pair, err := exchange.DecodeToPair(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = pair
		return contract, err
	}
	return nil, errors.New("decoding failure")
}
