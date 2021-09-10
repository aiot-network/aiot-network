package exchange_func

import (
	"errors"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/tools/arry"
)

type ExchangeAddLiquidity struct {
	Exchange       arry.Address
	TokenA         arry.Address
	TokenB         arry.Address
	To             arry.Address
	AmountADesired uint64
	AmountBDesired uint64
	AmountAMin     uint64
	AmountBMin     uint64
	Deadline       uint64
}

func (e *ExchangeAddLiquidity) Verify() error {
	if e.TokenA.IsEqual(e.TokenB) {
		return errors.New("invalid pair")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.TokenA.String()); !ok {
		return errors.New("wrong tokenA address")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.TokenB.String()); !ok {
		return errors.New("wrong tokenB address")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.Exchange.String()); !ok {
		return errors.New("wrong exchange address")
	}
	if ok := kit.CheckAddress(config.Param.Name, e.To.String()); !ok {
		return errors.New("wrong to address")
	}
	if e.AmountADesired == 0 {
		return errors.New("wrong amountADesired")
	}
	if e.AmountBDesired == 0 {
		return errors.New("wrong amountBDesired")
	}
	if e.AmountAMin > e.AmountADesired {
		return errors.New("wrong amountAMin")
	}
	if e.AmountBMin > e.AmountBDesired {
		return errors.New("wrong amountBMin")
	}
	return nil
}

type ExchangeRemoveLiquidity struct {
	Exchange  arry.Address
	TokenA    arry.Address
	TokenB    arry.Address
	To        arry.Address
	Liquidity  uint64
	AmountAMin uint64
	AmountBMin uint64
	Deadline   uint64
}

func (e *ExchangeRemoveLiquidity) Verify() error {
	if e.TokenA.IsEqual(e.TokenB) {
		return errors.New("invalid pair")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.TokenA.String()); !ok {
		return errors.New("wrong tokenA address")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.TokenB.String()); !ok {
		return errors.New("wrong tokenB address")
	}
	if ok := kit.CheckContractAddress(config.Param.Name, e.Exchange.String()); !ok {
		return errors.New("wrong exchange address")
	}
	if ok := kit.CheckAddress(config.Param.Name, e.To.String()); !ok {
		return errors.New("wrong to address")
	}
	return nil
}
