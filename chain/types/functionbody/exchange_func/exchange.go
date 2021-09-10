package exchange_func

import (
	"errors"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/tools/arry"
)

type ExchangeInitBody struct {
	Admin  arry.Address
	FeeTo  arry.Address
	Symbol string
}

func (e *ExchangeInitBody) Verify() error {
	if err := kit.CheckSymbol(e.Symbol); err != nil {
		return err
	}
	if ok := kit.CheckAddress(config.Param.Name, e.Admin.String()); !ok {
		return errors.New("wrong admin address")
	}
	feeTo := e.FeeTo.String()
	if feeTo != "" {
		if ok := kit.CheckAddress(config.Param.Name, feeTo); !ok {
			return errors.New("wrong feeTo address")
		}
	}
	return nil
}

type ExchangeAdmin struct {
	Address arry.Address
}

func (e *ExchangeAdmin) Verify() error {
	if ok := kit.CheckAddress(config.Param.Name, e.Address.String()); !ok {
		return errors.New("wrong admin address")
	}
	return nil
}

type ExchangeFeeTo struct {
	Address arry.Address
}

func (e *ExchangeFeeTo) Verify() error {
	if ok := kit.CheckAddress(config.Param.Name, e.Address.String()); !ok {
		return errors.New("wrong feeTo address")
	}
	return nil
}
