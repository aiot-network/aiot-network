package exchange_func

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/common/config"

	"github.com/aiot-network/aiotchain/tools/arry"
)

type ExactIn struct {
	AmountIn     uint64
	AmountOutMin uint64
	Path         []arry.Address
	To           arry.Address
	Deadline     uint64
}

func (e *ExactIn) Verify() error {
	if !kit.CheckAddress(config.Param.Name, e.To.String()) {
		return errors.New("wrong to address")
	}
	for _, addr := range e.Path {
		if !kit.CheckContractAddress(config.Param.Name, addr.String()) {
			return fmt.Errorf("wrong path address %s", addr.String())
		}
	}
	return nil
}

type ExactOut struct {
	AmountOut   uint64
	AmountInMax uint64
	Path        []arry.Address
	To          arry.Address
	Deadline    uint64
}

func (e *ExactOut) Verify() error {
	if !kit.CheckAddress(config.Param.Name, e.To.String()) {
		return errors.New("wrong to address")
	}
	for _, addr := range e.Path {
		if !kit.CheckContractAddress(config.Param.Name, addr.String()) {
			return fmt.Errorf("wrong path address %s", addr.String())
		}
	}
	return nil
}
