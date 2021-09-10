package runner

import (
	"fmt"

	"reflect"
	"strconv"
	"sync"
)

type ContractRunner struct {
	mutex   sync.RWMutex
	library *library.RunnerLibrary
}

func NewContractRunner(accountState _interface.IAccountState, contractState _interface.IContractState) *ContractRunner {
	library := library.NewRunnerLibrary(accountState, contractState)
	return &ContractRunner{
		library: library,
	}
}

func (c *ContractRunner) Verify(tx types.ITransaction, lastHeight uint64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if tx.GetTxType() != types.ContractV2_ {
		return nil
	}
	body, _ := tx.GetTxBody().(*types.TxContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library, tx, lastHeight)
		switch body.FunctionType {
		case contractv2.Exchange_Init:
			return ex.PreInitVerify()
		case contractv2.Exchange_SetAdmin:
			return ex.PreSetVerify()
		case contractv2.Exchange_SetFeeTo:
			return ex.PreSetVerify()
		case contractv2.Exchange_ExactIn:
			return ex.PreExactInVerify(lastHeight)
		case contractv2.Exchange_ExactOut:
			return ex.PreExactOutVerify(lastHeight)
		}

	case contractv2.Pair_:
		switch body.FunctionType {
		case contractv2.Pair_AddLiquidity:
			pair := exchange_runner.NewPairRunner(c.library, tx, 0, 0)
			return pair.PreAddLiquidityVerify()
		case contractv2.Pair_RemoveLiquidity:
			pair := exchange_runner.NewPairRunner(c.library, tx, 0, 0)
			return pair.PreRemoveLiquidityVerify(lastHeight)
		}
	}
	return nil
}

func (c *ContractRunner) RunContract(tx types.ITransaction, blockHeight uint64, blockTime uint64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	body, _ := tx.GetTxBody().(*types.TxContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library, tx, blockHeight)
		switch body.FunctionType {
		case contractv2.Exchange_Init:
			ex.Init()
		case contractv2.Exchange_SetAdmin:
			ex.SetAdmin()
		case contractv2.Exchange_SetFeeTo:
			ex.SetFeeTo()
		case contractv2.Exchange_ExactIn:
			ex.SwapExactIn(blockTime)
		case contractv2.Exchange_ExactOut:
			ex.SwapExactOut(blockTime)
		}
	case contractv2.Pair_:
		switch body.FunctionType {
		case contractv2.Pair_AddLiquidity:
			pairRunner := exchange_runner.NewPairRunner(c.library, tx, blockHeight, blockTime)
			pairRunner.AddLiquidity()
		case contractv2.Pair_RemoveLiquidity:
			pairRunner := exchange_runner.NewPairRunner(c.library, tx, blockHeight, blockTime)
			pairRunner.RemoveLiquidity()
		}
	}
	return nil
}

type Pair struct {
	Address  string  `json:"address"`
	Token0   string  `json:"token0"`
	Token1   string  `json:"token1"`
	Reserve0 float64 `json:"reserve0"`
	Reserve1 float64 `json:"reserve1"`
}

type openContract interface {
	Methods() map[string]*exchange_runner.MethodInfo
	MethodExist(method string) bool
}

func (c *ContractRunner) ReadMethod(address, method string, params []string) (interface{}, error) {
	var open openContract
	var err error
	contract := c.library.GetContractV2(address)
	if contract == nil {
		return nil, fmt.Errorf("contract %s dose not exist", address)
	}
	switch contract.Type {
	case contractv2.Exchange_:
		open, err = exchange_runner.NewExchangeState(c.library, address)
		if err != nil {
			return nil, err
		}
	case contractv2.Pair_:
		open, err = exchange_runner.NewPairState(c.library, address)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("method %s does not exist", method)
	}
	exist := open.MethodExist(method)
	if !exist {
		return nil, fmt.Errorf("method %s does not exist", method)
	}
	t := reflect.TypeOf(open)
	tMethod, exist := t.MethodByName(method)
	if !exist {
		return nil, fmt.Errorf("method %s does not exist", method)
	}
	inCount := tMethod.Type.NumIn()
	if inCount != len(params)+1 {
		return nil, fmt.Errorf("the number of parameters is %d", inCount-1)
	}
	interParams := make([]interface{}, inCount-1)
	for i := 1; i < tMethod.Type.NumIn(); i++ {
		paramT := tMethod.Type.In(i)
		switch paramT.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			iParam, err := strconv.ParseUint(params[i-1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parameter %d %s", i-1, err.Error())
			}
			interParams[i-1] = iParam
		case reflect.Float32, reflect.Float64:
			fParam, err := strconv.ParseFloat(params[i-1], 64)
			if err != nil {
				return nil, fmt.Errorf("parameter %d %s", i-1, err.Error())
			}
			interParams[i-1] = fParam
		case reflect.String:
			interParams[i-1] = params[i-1]
		case reflect.Bool:
			bParam, err := strconv.ParseBool(params[i-1])
			if err != nil {
				return nil, fmt.Errorf("parameter %d %s", i-1, err.Error())
			}
			interParams[i-1] = bParam
		case reflect.Ptr, reflect.Map, reflect.Array, reflect.Slice:
			interParams[i-1] = params[i-1]
		default:
			return nil, fmt.Errorf("parameter %d value type error", i-1)
		}

	}
	args := []reflect.Value{reflect.ValueOf(open)}
	for _, param := range interParams {
		args = append(args, reflect.ValueOf(param))
	}
	rs := tMethod.Func.Call(args)
	if len(rs) == 0 || len(rs) > 2 {
		return nil, fmt.Errorf("%s is not a read method", method)
	}
	if len(rs) == 2 {
		err, _ = rs[1].Interface().(error)
	}
	return reflectValue(rs[0]), err
}

func reflectValue(value reflect.Value) interface{} {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return value.Bool()
	case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Array, reflect.Slice:
		return value.Interface()
	case reflect.Interface:
		err, ok := value.Interface().(error)
		if ok {
			return err.Error()
		}
	}
	return nil
}
